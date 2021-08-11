package providers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	pipelinesv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	knativeapis "knative.dev/pkg/apis"
	knativeapisduckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

func TestNewSlackMessageFromPipelineRun(t *testing.T) {
	for _, c := range []struct {
		name string
		pr   *pipelinesv1beta1.PipelineRun
		req  *slackPostMessageRequest
	}{
		{
			name: "Basic",
			pr: &pipelinesv1beta1.PipelineRun{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "foo",
					Namespace:   "bar",
					Annotations: map[string]string{},
				},
				Status: pipelinesv1beta1.PipelineRunStatus{
					Status: knativeapisduckv1beta1.Status{
						Conditions: []knativeapis.Condition{
							{
								Type:    knativeapis.ConditionSucceeded,
								Status:  corev1.ConditionTrue,
								Reason:  "Reason",
								Message: "Message",
							},
						},
					},
					PipelineRunStatusFields: pipelinesv1beta1.PipelineRunStatusFields{
						StartTime: &metav1.Time{
							Time: time.Date(2020, 6, 1, 17, 44, 13, 0, time.Local),
						},
						CompletionTime: &metav1.Time{
							Time: time.Date(2020, 6, 1, 18, 47, 18, 0, time.Local),
						},
						TaskRuns:        map[string]*pipelinesv1beta1.PipelineRunTaskRunStatus{},
						Runs:            map[string]*pipelinesv1beta1.PipelineRunRunStatus{},
						PipelineResults: []pipelinesv1beta1.PipelineRunResult{},
						PipelineSpec:    &pipelinesv1beta1.PipelineSpec{},
						SkippedTasks:    []pipelinesv1beta1.SkippedTask{},
					},
				},
			},
			req: &slackPostMessageRequest{
				Channel:  "",
				Fallback: "Reason: foo.bar",
				Attachments: []slackAttachment{
					{
						Color: slackColorGood,
						Blocks: []slackBlock{
							{
								Type: "section",
								Text: &slackBlockText{
									Type: "mrkdwn",
									Text: "*foo.bar*",
								},
							},
							{
								Type: "section",
								Text: &slackBlockText{
									Type: "plain_text",
									Text: "Reason: Message",
								},
							},
							{
								Type: "context",
								Elements: []slackBlockElement{
									{
										Type: "mrkdwn",
										Text: "1h3m5s",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "WithDashboardAnnotation",
			pr: &pipelinesv1beta1.PipelineRun{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
					Annotations: map[string]string{
						"integrations.tekton.ornew.io/tekton-dashboard-base-url": "http://example.com/",
					},
				},
				Status: pipelinesv1beta1.PipelineRunStatus{
					Status: knativeapisduckv1beta1.Status{
						Conditions: []knativeapis.Condition{
							{
								Type:    knativeapis.ConditionSucceeded,
								Status:  corev1.ConditionFalse,
								Reason:  "Reason",
								Message: "Message",
							},
						},
					},
					PipelineRunStatusFields: pipelinesv1beta1.PipelineRunStatusFields{
						StartTime: &metav1.Time{
							Time: time.Date(2020, 6, 1, 17, 44, 13, 0, time.Local),
						},
						CompletionTime: &metav1.Time{
							Time: time.Date(2020, 6, 1, 18, 47, 18, 0, time.Local),
						},
						TaskRuns:        map[string]*pipelinesv1beta1.PipelineRunTaskRunStatus{},
						Runs:            map[string]*pipelinesv1beta1.PipelineRunRunStatus{},
						PipelineResults: []pipelinesv1beta1.PipelineRunResult{},
						PipelineSpec:    &pipelinesv1beta1.PipelineSpec{},
						SkippedTasks:    []pipelinesv1beta1.SkippedTask{},
					},
				},
			},
			req: &slackPostMessageRequest{
				Channel:  "",
				Fallback: "Reason: foo.bar",
				Attachments: []slackAttachment{
					{
						Color: slackColorDanger,
						Blocks: []slackBlock{
							{
								Type: "section",
								Text: &slackBlockText{
									Type: "mrkdwn",
									Text: "*foo.bar*",
								},
							},
							{
								Type: "section",
								Text: &slackBlockText{
									Type: "plain_text",
									Text: "Reason: Message",
								},
							},
							{
								Type: "context",
								Elements: []slackBlockElement{
									{
										Type: "mrkdwn",
										Text: "1h3m5s | <http://example.com/#/namespaces/bar/pipelineruns/foo|open dashboard>",
									},
								},
							},
						},
					},
				},
			},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			act := newSlackMessageFromPipelineRun(c.pr)
			assert.NotNil(t, act)
			assert.Equal(t, c.req, act)
		})
	}
}
