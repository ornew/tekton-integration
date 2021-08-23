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
	"time"

	ce "github.com/cloudevents/sdk-go/v2"
	cebinding "github.com/cloudevents/sdk-go/v2/binding"
	ceprotocol "github.com/cloudevents/sdk-go/v2/protocol"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/google/uuid"
	"github.com/ornew/tekton-integration/pkg/api/v1alpha1"
	pipelinesv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CloudEventsProtocol interface {
	Send(ctx context.Context, e ce.Event) ceprotocol.Result
}

type CloudEvents struct {
	Protocol CloudEventsProtocol
}

var _ Provider = (*CloudEvents)(nil)

func (a *CloudEvents) Notify(ctx context.Context, pr *pipelinesv1beta1.PipelineRun) error {
	event, err := a.toEvent(pr)
	if err != nil {
		return err
	}
	if result := a.Protocol.Send(ctx, *event); !ce.IsACK(result) {
		return NewRuntimeError(fmt.Sprintf("failed to send, %v", result))
	}
	return nil
}

func (a *CloudEvents) toEvent(pr *pipelinesv1beta1.PipelineRun) (*ce.Event, error) {
	event := ce.NewEvent()
	// TODO
	event.SetSource("example/uri")
	event.SetType("example.type")
	event.SetID(uuid.New().String())
	event.SetTime(time.Now())
	event.SetData(ce.ApplicationJSON, pr)
	if err := event.Validate(); err != nil {
		return nil, NewRuntimeError(fmt.Sprintf("validation failed: %v", err))
	}
	return &event, nil
}

func NewCloudEvents(ctx context.Context, p *v1alpha1.Provider, k client.Client) (*CloudEvents, error) {
	s := p.Spec.CloudEvents
	if s == nil {
		return nil, NewInvalidProviderSpecError("missing value .cloudEvents")
	}
	switch s.Protocol {
	case "Webhook":
		if s.Webhook == nil {
			return nil, NewInvalidProviderSpecError("missing value .cloudEvents.webhook")
		}
		// TODO validate sink url
		// TODO if authorization, override transport by WithRoundTripper
		c, err := cehttp.New(
		//cehttp.WithRoundTripper(),
		)
		if err != nil {
			return nil, NewRuntimeError(fmt.Sprintf("failed to create http client, %v", err))
		}
		// TODO if validation, check OPTIONS
		return &CloudEvents{
			Protocol: &CloudEventsWebhook{
				SinkURL: s.Webhook.URL,
				Sender:  c,
			},
		}, nil
	default:
		return nil, NewInvalidProviderSpecError(fmt.Sprintf("unknown CloudEvents protocol: %s", s.Protocol))
	}
}

type CloudEventsWebhook struct {
	SinkURL string
	Sender  ceprotocol.Sender
}

func (c *CloudEventsWebhook) Send(ctx context.Context, e ce.Event) ceprotocol.Result {
	ctx = ce.ContextWithTarget(ctx, c.SinkURL)
	return c.Sender.Send(ctx, (*cebinding.EventMessage)(&e))
}
