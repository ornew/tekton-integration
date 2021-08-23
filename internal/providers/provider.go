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
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/client"

	pipelinesv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"

	"github.com/ornew/tekton-integration/pkg/api/v1alpha1"
)

const (
	annotationContextID              = "integrations.tekton.ornew.io/context-id"
	annotationTektonDashboardBaseURL = "integrations.tekton.ornew.io/tekton-dashboard-base-url"
)

type Provider interface {
	Notify(ctx context.Context, pr *pipelinesv1beta1.PipelineRun) error
}

func ResolveProvider(ctx context.Context, p *v1alpha1.Provider, k8s client.Client) (app Provider, err error) {
	switch p.Spec.Type {
	case "GitHubApp":
		app, err = NewGitHubApp(ctx, p, k8s)
		return
	case "SlackApp":
		app, err = NewSlackApp(ctx, p, k8s)
		return
	}
	return nil, NewInvalidProviderSpecError(fmt.Sprintf("unknown provider type: %v", p.Spec.Type))
}

func getDashboardPipelineRunURL(base, namespace, name string) string {
	return fmt.Sprintf("%s/#/namespaces/%s/pipelineruns/%s", strings.TrimSuffix(base, "/"), namespace, name)
}
