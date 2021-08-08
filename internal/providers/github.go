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

type GitHubApp struct {
	AppId      int64
	PrivateKey SecretBytes
	BaseURL    *string
}

var _ Provider = (*GitHubApp)(nil)

func NewGitHubApp(ctx context.Context, p *v1alpha1.Provider, k client.Client) (*GitHubApp, error) {
	s := p.Spec.GitHubApp
	if s == nil {
		return nil, NewInvalidProviderSpecError("missing value .githubApp")
	}
	var key []byte
	if s.PrivateKey.SecretRef != nil {
		var secret corev1.Secret
		ref := types.NamespacedName{
			Namespace: p.Namespace,
			Name:      s.PrivateKey.SecretRef.Name,
		}
		if err := k.Get(ctx, ref, &secret); err != nil {
			return nil, NewNotFoundPrivateKeyError(fmt.Sprintf("failed to get secret: %v", err))
		}
		if secret.Data == nil {
			return nil, NewNotFoundPrivateKeyError("data not found in secret")
		}
		if pem, ok := secret.Data["private-key.pem"]; ok {
			key = pem
		} else {
			return nil, NewNotFoundPrivateKeyError("missing key private-key.pem")
		}
	} else {
		return nil, NewInvalidProviderSpecError("missing valid values in .privateKey")
	}
	return &GitHubApp{
		AppId:      s.AppId,
		PrivateKey: key,
		BaseURL:    s.BaseURL,
	}, nil
}

func (a *GitHubApp) Notify(ctx context.Context, pr *pipelinesv1beta1.PipelineRun) *ProviderError {
	log := logr.FromContext(ctx).WithName("providers.githubapp").
		WithValues("providerType", "GitHubApp", "pipelineRun", pr.Name)
	contextID := pr.Annotations[annotationContextID]
	if len(contextID) < 1 {
		ref := pr.Spec.PipelineRef
		if ref != nil && len(ref.Name) > 0 {
			contextID = ref.Name
		}
		return NewFailedValidationError("context-id or pipelineRef.name is required")
	}
	owner := pr.Annotations[annotationGitHubOwner]
	repo := pr.Annotations[annotationGitHubRepo]
	revision := pr.Annotations[annotationGitHubSHA]
	if len(owner) < 1 || len(repo) < 1 || len(revision) < 1 {
		return NewFailedValidationError(fmt.Sprintf("required annotations: %s=%s %s=%s %s=%s",
			annotationGitHubOwner, owner,
			annotationGitHubRepo, repo,
			annotationGitHubSHA, revision,
		))
	}
	cond := pr.Status.GetCondition(apis.ConditionSucceeded)
	if cond == nil {
		log.V(1).Info("PipelineRun has not condition, ignored")
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
		return NewRuntimeError(fmt.Sprintf("failed to get GitHub App transport: %v", err))
	}
	if a.BaseURL != nil {
		atr.BaseURL = *a.BaseURL
	}
	bearerClient := github.NewClient(&http.Client{Transport: atr})
	ins, _, err := bearerClient.Apps.FindRepositoryInstallation(ctx, owner, repo)
	if err != nil || ins.ID == nil {
		return NewRuntimeError(fmt.Sprintf("failed to find GitHub App installation: %v", err))
	}
	itr := ghinstallation.NewFromAppsTransport(atr, *ins.ID)

	var client *github.Client
	if a.BaseURL != nil {
		client, err = github.NewEnterpriseClient(itr.BaseURL, itr.BaseURL, &http.Client{Transport: itr})
		if err != nil {
			return NewRuntimeError(fmt.Sprintf("failed to get GitHub Enterprise API client: %v", err))
		}
	} else {
		client = github.NewClient(&http.Client{Transport: itr})
	}
	status, _, err = client.Repositories.CreateStatus(ctx, owner, repo, revision, status)
	if err != nil {
		return NewRuntimeError(fmt.Sprintf("failed to set GitHub commit status: %v", err))
	}
	log.V(2).Info("set commit status", "status", status)
	return nil
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
