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
	"net/http"
	"net/http/httptest"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cebinding "github.com/cloudevents/sdk-go/v2/binding"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/ornew/tekton-integration/pkg/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pipelinesv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

func TestNewCloudEvents(t *testing.T) {
	for _, c := range []struct {
		p       *v1alpha1.Provider
		wantErr *ProviderError
	}{
		{
			p: &v1alpha1.Provider{
				TypeMeta: metav1.TypeMeta{},
				Spec: v1alpha1.ProviderSpec{
					Type: "CloudEvents",
					CloudEvents: &v1alpha1.CloudEventsSpec{
						Protocol: "Webhook",
						Webhook: &v1alpha1.CloudEventsWebhookSpec{
							URL: "",
							//Authorization: v1alpha1.HTTPAuthorization{},
						},
					},
				},
			},
		},
	} {
		ctx := context.TODO()
		_, err := NewCloudEvents(ctx, c.p, nil)
		if c.wantErr != nil {
			assert.Error(t, err)
			var perr *ProviderError
			if assert.ErrorAs(t, err, &perr) {
				assert.Equal(t, c.wantErr.Code, perr.Code)
			}
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestCloudEvents(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := cehttp.NewMessageFromHttpRequest(r)
		evt, err := cebinding.ToEvent(context.TODO(), m)
		m.Finish(err)
		assert.NoError(t, err)
		t.Log(evt)
		w.WriteHeader(http.StatusOK)
	})
	ts := httptest.NewServer(h)
	defer ts.Close()

	pr := &pipelinesv1beta1.PipelineRun{}

	p1 := &v1alpha1.Provider{
		Spec: v1alpha1.ProviderSpec{
			Type: "CloudEvents",
			CloudEvents: &v1alpha1.CloudEventsSpec{
				Protocol: "Webhook",
				Webhook: &v1alpha1.CloudEventsWebhookSpec{
					URL: ts.URL,
					//Authorization: v1alpha1.HTTPAuthorization{},
				},
			},
		},
	}
	ce1, perr := NewCloudEvents(ctx, p1, nil)
	require.NoError(t, perr)
	perr = ce1.Notify(context.TODO(), pr)
	assert.NoError(t, perr)
}
