# Tekton Integrations

Tekton Integrations provides a way to integrate Tekton with external services.

Use Cases:

- Sync the status of PipelineRun to the commit status on GitHub.
- Notifies the result of PipelineRun to Slack channels.

Tekton Integrations consists of CRD:

- `Provider` presents credentials and settings for integrating with external services.
- `Notification` presents how to notify the result of Run.

For example:

```yaml
apiVersion: integrations.tekton.ornew.io/v1alpha1
kind: Provider
metadata:
  name: github-app
spec:
  type: GitHubApp
  githubApp:
    appId: 1
    privateKey:
      secretRef:
        name: github-app
---
apiVersion: integrations.tekton.ornew.io/v1alpha1
kind: Notification
metadata:
  name: github-commit-statuses
spec:
  providerRef:
    name: github-app
```

- [Provider](docs/provider.md)
- [Notification](docs/notification.md)

## Alternatives

### CloudEvents + Tekton Triggers

Tekton supports publishing CloudEvents and Tekton Triggers can handle the events.
Certainly tektoncd/plumbing is handling CloudEvents to make up the pipeline.
However, the status of Run changes frequently, so a custom object is created
and the pod is launched each time. This not only uses the resources of
the cluster, but also makes it difficult to reuse connections and credentials.
Trigger is useful when events are infrequent, but expensive when connecting
frequently occurring events to a typical external service.

Since only one CloudEvents sink can be set by default, any PubSub is required
for some integrations. Tekton Integrations does not require this setup.

When Tekton Triggers handles CloudEvents published by Tekton,
recursive execution can occur and consume resources unlimitedly in the cluster.
You need to filter the events from the triggered Run, and set resource quotas
to block unexpected many Runs. Recursively Runs does not occur
in Tekton Integrations.

## Supported Providers

WIP

- [GitHub App](docs/providers/github.md)
- [Slack App](docs/providers/slack.md)
- AWS SNS
- GCP PubSub
- CloudEvents

## Setup

TBW
