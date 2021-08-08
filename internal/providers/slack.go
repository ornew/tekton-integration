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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	pipelinesv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"

	"github.com/ornew/tekton-integration/api/v1alpha1"
)

const (
	slackPostMessageURL = "https://slack.com/api/chat.postMessage"
)

type SlackApp struct {
	AccessToken SecretBytes
	Channels    []v1alpha1.SlackChannel
}

var _ Provider = (*SlackApp)(nil)

func NewSlackApp(ctx context.Context, p *v1alpha1.Provider, k client.Client) (*SlackApp, *ProviderError) {
	s := p.Spec.SlackApp
	if s == nil {
		return nil, NewInvalidProviderSpecError("missing value .slackApp")
	}
	var key []byte
	if s.AccessToken.SecretRef != nil {
		var secret corev1.Secret
		ref := types.NamespacedName{
			Namespace: p.Namespace,
			Name:      s.AccessToken.SecretRef.Name,
		}
		if err := k.Get(ctx, ref, &secret); err != nil {
			return nil, NewNotFoundPrivateKeyError(fmt.Sprintf("failed to get secret: %v", err))
		}
		if secret.Data == nil {
			return nil, NewNotFoundPrivateKeyError("data not found in secret")
		}
		if pem, ok := secret.Data["access-token"]; ok {
			key = pem
		} else {
			return nil, NewNotFoundPrivateKeyError("missing key access-token")
		}
	} else {
		return nil, NewInvalidProviderSpecError("missing valid values in .privateKey")
	}
	return &SlackApp{
		AccessToken: NewSecretBytes(key),
		Channels:    s.Channels,
	}, nil
}

func postHTTP(url string, auth string, payload interface{}) (*http.Response, error) {
	client := &http.Client{}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", auth)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (a *SlackApp) Notify(ctx context.Context, pr *pipelinesv1beta1.PipelineRun) *ProviderError {
	log := logr.FromContext(ctx).WithName("providers.slackapp").
		WithValues("providerType", "SlackApp", "pipelineRun", pr.Name)
	cond := pr.Status.GetCondition(apis.ConditionSucceeded)
	switch cond.Status {
	case corev1.ConditionTrue:
	case corev1.ConditionFalse:
	case corev1.ConditionUnknown:
		log.V(2).Info("this run is not finished yet, skipped")
		return nil
	}
	for _, channel := range a.Channels {
		c, perr := resolveSlackChannel(channel)
		if perr != nil {
			return perr
		}
		payload := newSlackMessageFromPipelineRun(pr)
		payload.Channel = c
		log.V(2).Info("payload", "payload", payload)
		bearer := fmt.Sprintf("Bearer %s", a.AccessToken.GetNoRedactedString())
		resp, err := postHTTP(slackPostMessageURL, bearer, payload)
		if err != nil {
			return NewRuntimeError(fmt.Sprintf("failed to post Slack message: %v", err))
		}
		b, err := ioutil.ReadAll(resp.Body)
		var r slackPostMessageResponse
		if err = json.Unmarshal(b, &r); err != nil {
			return NewRuntimeError(fmt.Sprintf("failed to unmarshal Slack response message: %v", err))
		}
		if !r.OK {
			errm := ""
			if r.Error != nil {
				errm = *r.Error
			}
			return NewRuntimeError(fmt.Sprintf("get an error from Slack: %s", errm))
		}
		log.V(2).Info("post message", "response", r)
	}
	return nil
}

type slackBlockText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type slackBlockElement struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type slackBlock struct {
	Type     string              `json:"type"`
	Text     *slackBlockText     `json:"text,omitempty"`
	Elements []slackBlockElement `json:"elements,omitempty"`
}

type slackAttachment struct {
	Color  string       `json:"color"`
	Blocks []slackBlock `json:"blocks"`
}

type slackPostMessageRequest struct {
	Channel     string            `json:"channel"`
	Fallback    string            `json:"fallback"`
	Attachments []slackAttachment `json:"attachments"`
}

type slackPostMessageResponse struct {
	OK        bool    `json:"ok"`
	Channel   *string `json:"channel,omitempty"`
	Timestamp *string `json:"ts,omitempty"`
	Error     *string `json:"error,omitempty"`
}

func newSlackMessageFromPipelineRun(pr *pipelinesv1beta1.PipelineRun) *slackPostMessageRequest {
	cond := pr.Status.GetCondition(apis.ConditionSucceeded)
	nn := fmt.Sprintf("%s.%s", pr.Name, pr.Namespace)
	reason := cond.Reason
	message := cond.Message
	color := getSlackColor(cond)
	duration := pr.Status.CompletionTime.Time.Sub(pr.Status.StartTime.Time)
	context := strings.Builder{}
	fmt.Fprint(&context, duration)
	dashboardBaseURL := pr.Annotations[annotationTektonDashboardBaseURL]
	if len(dashboardBaseURL) > 0 {
		url := getDashboardPipelineRunURL(dashboardBaseURL, pr.Namespace, pr.Name)
		fmt.Fprintf(&context, "| <%s|open dashboard>", url)
	}
	return &slackPostMessageRequest{
		Fallback: fmt.Sprintf("%s: %s", reason, nn),
		Attachments: []slackAttachment{
			{
				Color: color,
				Blocks: []slackBlock{
					{
						Type: "section",
						Text: &slackBlockText{
							Type: "mrkdwn",
							Text: fmt.Sprintf("*%s*", nn),
						},
					},
					{
						Type: "section",
						Text: &slackBlockText{
							Type: "plain_text",
							Text: fmt.Sprintf("%s: %s", reason, message),
						},
					},
					{
						Type: "context",
						Elements: []slackBlockElement{
							{
								Type: "mrkdwn",
								Text: context.String(),
							},
						},
					},
				},
			},
		},
	}
}

const (
	slackColorGood    = "#2EB886"
	slackColorWarning = "#DAA038"
	slackColorDanger  = "#A30100"
)

func getSlackColor(c *apis.Condition) string {
	switch c.Status {
	case corev1.ConditionUnknown:
		return slackColorWarning
	case corev1.ConditionTrue:
		return slackColorGood
	case corev1.ConditionFalse:
		return slackColorDanger
	}
	return slackColorDanger
}

func resolveSlackChannel(c v1alpha1.SlackChannel) (string, *ProviderError) {
	if c.ID != nil {
		return *c.ID, nil
	}
	if c.Name != nil {
		return *c.Name, nil
	}
	return "", NewInvalidProviderSpecError("Slack channel id or name is required")
}
