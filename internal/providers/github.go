/*
Copyright 2021 Arata Furukawa.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package providers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/go-logr/logr"
	"github.com/google/go-github/v37/github"
	pipelinesv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"

	"github.com/ornew/tekton-integration/api/v1alpha1"
)

const (
	annotationGitHubOwner = "integrations.tekton.ornew.io/github-owner"
	annotationGitHubRepo  = "integrations.tekton.ornew.io/github-repo"
	annotationGitHubSHA   = "integrations.tekton.ornew.io/github-sha"
)

var (
	errGitHubAppInvalidSpec       = errors.New("invalid spec")
	errGitHubAppFailedToGetSecret = errors.New("failed to get secret")
	errGitHubAppMissingPrivateKey = errors.New("missing private-key.pem")
)

type GitHubApp struct {
	AppId      int64
	PrivateKey []byte
	BaseURL    *string
}

func NewGitHubApp(ctx context.Context, p *v1alpha1.Provider, k client.Client) (*GitHubApp, error) {
	s := p.Spec.GitHubApp
	if s == nil {
		return nil, NewGitHubAppError(ErrorCodeGitHubAppInvalidSpec, "missing .githubApp")
	}
	var key []byte
	if s.PrivateKey.SecretRef != nil {
		var secret corev1.Secret
		ref := types.NamespacedName{
			Namespace: p.Namespace,
			Name:      s.PrivateKey.SecretRef.Name,
		}
		if err := k.Get(ctx, ref, &secret); err != nil {
			return nil, NewGitHubAppError(ErrorCodeGitHubAppPrivateKeyNotFound, fmt.Sprintf("failed to get secret: %v", err))
		}
		if secret.Data == nil {
			return nil, NewGitHubAppError(ErrorCodeGitHubAppPrivateKeyNotFound, "data not found in secret")
		}
		if pem, ok := secret.Data["private-key.pem"]; ok {
			key = pem
		} else {
			return nil, NewGitHubAppError(ErrorCodeGitHubAppPrivateKeyNotFound, "missing key private-key.pem")
		}
	} else {
		return nil, NewGitHubAppError(ErrorCodeGitHubAppInvalidSpec, "missing valid .privateKey")
	}
	return &GitHubApp{
		AppId:      s.AppId,
		PrivateKey: key,
		BaseURL:    s.BaseURL,
	}, nil
}

func (a *GitHubApp) Notify(ctx context.Context, pr *pipelinesv1beta1.PipelineRun) error {
	log := logr.FromContext(ctx).WithName("provider.githubapp").WithValues("providerType", "GitHubApp", "pipelinerun", pr.Name)
	log.Info("notifications github")
	contextID := pr.Annotations[annotationContextID]
	if len(contextID) < 1 {
		ref := pr.Spec.PipelineRef
		if ref != nil && len(ref.Name) > 0 {
			contextID = ref.Name
		}
		return fmt.Errorf("context id is not found")
	}
	owner := pr.Annotations[annotationGitHubOwner]
	repo := pr.Annotations[annotationGitHubRepo]
	revision := pr.Annotations[annotationGitHubSHA]
	if len(owner) < 1 || len(repo) < 1 || len(revision) < 1 {
		err := fmt.Errorf("required annotations: %s=%s %s=%s %s=%s",
			annotationGitHubOwner, owner,
			annotationGitHubRepo, repo,
			annotationGitHubSHA, revision,
		)
		log.Error(err, "missing annotations")
		return err
	}
	cond := pr.Status.GetCondition(apis.ConditionSucceeded)
	if cond == nil {
		log.Info("PipelineRun has not condition, ignored")
		return nil
	}
	state := toGithubCommitStatus(cond.Status)
	description := cond.Reason
	context := fmt.Sprintf("tekton: %s", contextID)
	targetURL := "" // TODO will support tekton dashboard or custom link
	status := &github.RepoStatus{
		State:       &state, // pending, success, error, or failure
		TargetURL:   &targetURL,
		Description: &description, // max len 140
		Context:     &context,
	}

	tr := http.DefaultTransport
	atr, err := ghinstallation.NewAppsTransport(tr, a.AppId, a.PrivateKey)
	if err != nil {
		log.Error(err, "failed to get GitHub App transport")
		return err
	}
	if a.BaseURL != nil {
		atr.BaseURL = *a.BaseURL
	}
	bearerClient := github.NewClient(&http.Client{Transport: atr})
	ins, _, err := bearerClient.Apps.FindRepositoryInstallation(ctx, owner, repo)
	if err != nil || ins.ID == nil {
		log.Error(err, "failed to get GitHub App installation")
		return err
	}
	itr := ghinstallation.NewFromAppsTransport(atr, *ins.ID)

	var client *github.Client
	if a.BaseURL != nil {
		client, err = github.NewEnterpriseClient(itr.BaseURL, itr.BaseURL, &http.Client{Transport: itr})
		if err != nil {
			log.Error(err, "failed to get GitHub Enterprise API client")
			return err
		}
	} else {
		client = github.NewClient(&http.Client{Transport: itr})
	}
	status, _, err = client.Repositories.CreateStatus(ctx, owner, repo, revision, status)
	if err != nil {
		log.Error(err, "failed to set GitHub commit status")
		return err
	}
	log.Info("set commit status", "status", status)
	return nil
}

// NOTE maybe providers can have a common error type
type GitHubAppErrorCode string

const (
	ErrorCodeGitHubAppInvalidSpec        = GitHubAppErrorCode("InvalidSpec")
	ErrorCodeGitHubAppPrivateKeyNotFound = GitHubAppErrorCode("PrivateKeyNotFound")
)

type GitHubAppError struct {
	Code    GitHubAppErrorCode
	Message string
}

func (e *GitHubAppError) Error() string {
	return fmt.Sprintf("%s (code=%s)", e.Message, e.Code)
}

func NewGitHubAppError(code GitHubAppErrorCode, msg string) error {
	return &GitHubAppError{
		Code:    code,
		Message: msg,
	}
}

const (
	GitHubCommitStatusPending    = "pending"
	GitHubCommitStatusSuccessful = "success"
	GitHubCommitStatusError      = "error"
	GitHubCommitStatusFailure    = "failure"
)

func toGithubCommitStatus(status corev1.ConditionStatus) string {
	switch status {
	case corev1.ConditionUnknown:
		return GitHubCommitStatusPending
	case corev1.ConditionTrue:
		return GitHubCommitStatusSuccessful
	case corev1.ConditionFalse:
		return GitHubCommitStatusError
	}
	return GitHubCommitStatusFailure
}
