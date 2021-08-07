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
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/go-logr/logr"
	"github.com/google/go-github/v37/github"
	pipelinesv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"

	"github.com/ornew/tekton-integration/api/v1alpha1"
)

const (
	GitHubCommitStatusPending    = "pending"
	GitHubCommitStatusFailed     = "failed"
	GitHubCommitStatusSuccessful = "successful"
)

func toGithubCommitStatus(status corev1.ConditionStatus) string {
	switch status {
	case corev1.ConditionUnknown:
		return GitHubCommitStatusPending
	case corev1.ConditionFalse:
		return GitHubCommitStatusFailed
	case corev1.ConditionTrue:
		return GitHubCommitStatusSuccessful
	}
	return GitHubCommitStatusPending
}

type Provider interface {
	Notify(ctx context.Context, pr *pipelinesv1beta1.PipelineRun) error
}

type GitHubApp struct {
	AppId      int64
	PrivateKey []byte
	BaseURL    *string
}

func NewGitHubApp(ctx context.Context, p *v1alpha1.Provider, k client.Client) (*GitHubApp, error) {
	s := p.Spec.GitHubApp
	if s == nil {
		return nil, errors.New(".spec.githubApp is required")
	}
	var key []byte
	if s.PrivateKey.SecretRef != nil {
		var secret corev1.Secret
		ref := types.NamespacedName{
			Namespace: p.Namespace,
			Name:      s.PrivateKey.SecretRef.Name,
		}
		if err := k.Get(ctx, ref, &secret); err != nil {
			return nil, fmt.Errorf("failed to get secret: %v", ref)
		}
		if secret.Data == nil {
			return nil, fmt.Errorf("failed to get private key because data is nothing: %v", ref)
		}
		if pem, ok := secret.Data["private-key.pem"]; ok {
			key = pem
		} else {
			return nil, errors.New("private-key.pem is not found in the secret")
		}
	} else {
		return nil, errors.New("secretRef is missing")
	}
	return &GitHubApp{
		AppId:      s.AppId,
		PrivateKey: key,
		BaseURL:    s.BaseURL,
	}, nil
}

func (a *GitHubApp) Notify(ctx context.Context, pr *pipelinesv1beta1.PipelineRun) error {
	log := logr.FromContext(ctx)
	log.Info("notifications github")
	owner := ""
	repo := ""
	revision := ""

	if len(owner) < 0 {
		return nil // FIXME TODO
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
	status := &github.RepoStatus{
		State:       new(string), // pending, success, error, or failure
		TargetURL:   new(string),
		Description: new(string), // max len 140
		Context:     new(string),
	}
	client.Repositories.CreateStatus(ctx, owner, repo, revision, status)

	return nil
}
