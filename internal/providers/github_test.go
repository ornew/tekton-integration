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
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	pipelinesv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"

	"github.com/ornew/tekton-integration/api/v1alpha1"
)

var (
	ctx context.Context = context.TODO()
)

func TestNewGitHubApp(t *testing.T) {
	kb := fakeclient.NewClientBuilder().
		WithObjects(
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"private-key.pem": []byte("private-key"),
				},
			},
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret-missing-private-key",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"foo": []byte(""),
				},
			},
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret-missing-data",
					Namespace: "default",
				},
				Data: nil,
			})
	k := kb.Build()
	for _, c := range []struct {
		name     string
		provider *v1alpha1.Provider
		wantErr  *ProviderError
	}{
		{
			name: "Basic",
			provider: &v1alpha1.Provider{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "provider",
					Namespace: "default",
				},
				Spec: v1alpha1.ProviderSpec{
					Type: "GitHubApp",
					GitHubApp: &v1alpha1.GitHubAppSpec{
						AppId: 1,
						PrivateKey: v1alpha1.PrivateKeySource{
							SecretRef: &v1alpha1.LocalSecretKeyReference{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "secret",
								},
							},
						},
						BaseURL: pointer.String("https://github.enterpise"),
					},
				},
				Status: v1alpha1.ProviderStatus{},
			},
		},
		{
			name: "GitHubAppSpecNotFound",
			provider: &v1alpha1.Provider{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "provider",
					Namespace: "default",
				},
				Spec: v1alpha1.ProviderSpec{
					Type:      "GitHubApp",
					GitHubApp: nil,
				},
				Status: v1alpha1.ProviderStatus{},
			},
			wantErr: NewInvalidProviderSpecError(""),
		},
		{
			name: "ValidPrivateKeyNotFound",
			provider: &v1alpha1.Provider{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "provider",
					Namespace: "default",
				},
				Spec: v1alpha1.ProviderSpec{
					Type: "GitHubApp",
					GitHubApp: &v1alpha1.GitHubAppSpec{
						AppId: 1,
						PrivateKey: v1alpha1.PrivateKeySource{
							SecretRef: nil,
						},
						BaseURL: pointer.String("https://github.enterpise"),
					},
				},
				Status: v1alpha1.ProviderStatus{},
			},
			wantErr: NewInvalidProviderSpecError(""),
		},
		{
			name: "SecretNotFound",
			provider: &v1alpha1.Provider{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "provider",
					Namespace: "default",
				},
				Spec: v1alpha1.ProviderSpec{
					Type: "GitHubApp",
					GitHubApp: &v1alpha1.GitHubAppSpec{
						AppId: 1,
						PrivateKey: v1alpha1.PrivateKeySource{
							SecretRef: &v1alpha1.LocalSecretKeyReference{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "not-exists-secret",
								},
							},
						},
					},
				},
				Status: v1alpha1.ProviderStatus{},
			},
			wantErr: NewNotFoundPrivateKeyError(""),
		},
		{
			name: "SecretPrivateKeyNotFound",
			provider: &v1alpha1.Provider{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "provider",
					Namespace: "default",
				},
				Spec: v1alpha1.ProviderSpec{
					Type: "GitHubApp",
					GitHubApp: &v1alpha1.GitHubAppSpec{
						AppId: 1,
						PrivateKey: v1alpha1.PrivateKeySource{
							SecretRef: &v1alpha1.LocalSecretKeyReference{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "secret-missing-private-key",
								},
							},
						},
					},
				},
				Status: v1alpha1.ProviderStatus{},
			},
			wantErr: NewNotFoundPrivateKeyError(""),
		},
		{
			name: "SecretDataNotFound",
			provider: &v1alpha1.Provider{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "provider",
					Namespace: "default",
				},
				Spec: v1alpha1.ProviderSpec{
					Type: "GitHubApp",
					GitHubApp: &v1alpha1.GitHubAppSpec{
						AppId: 1,
						PrivateKey: v1alpha1.PrivateKeySource{
							SecretRef: &v1alpha1.LocalSecretKeyReference{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "secret-missing-data",
								},
							},
						},
					},
				},
				Status: v1alpha1.ProviderStatus{},
			},
			wantErr: NewNotFoundPrivateKeyError(""),
		},
	} {
		c := c
		t.Run(c.name, func(t *testing.T) {
			a, err := NewGitHubApp(ctx, c.provider, k)
			if c.wantErr != nil {
				assert.Error(t, err)
				var terr *ProviderError
				if assert.ErrorAs(t, err, &terr) {
					// check error code only in this test, ignore message
					assert.Equal(t, c.wantErr.Code, terr.Code)
				}
			} else {
				assert.Nil(t, err)
				assert.Equal(t, a.AppId, int64(1))
				assert.Equal(t, a.PrivateKey.GetNoRedactedString(), "private-key")
				assert.Equal(t, *a.BaseURL, "https://github.enterpise")
			}
		})
	}
}

func TestGitHubAppNotify(t *testing.T) {
	_ = &pipelinesv1beta1.PipelineRun{}
	// TODO
}
