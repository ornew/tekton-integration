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

Tekton publishes CloudEvents of Runs. You can also use Tekton Triggers to handle
CloudEvents. This configuration is actually used in tektoncd/plumbing.
Not only does this consume resources linearly, but those pods incur unnecessary
overhead due to the difficulty of sharing credentials and connections.
As Tekton's workload grows, API rate limiting and the number of network
connections can be problematic.
With Tekton Integration, you can cache your credentials on the controller to
prevent it from making more connections than you need.

Another reason is that recursive execution may occur.
If Tekton publishes the status of the Run to CloudEvents and Trigger occurred,
you can see that the triggered Run also raises CloudEvents.
Users are need to handle events carefully, such as annotating Runs resulting
from Trigger and filtering with CEL expressions.
In the unlikely event of recursion without properly setting resource quotas,
the entire cluster will suffer a major failure. In my experience,
the etcd of API server will be down, leaving the cluster out of control.
Tekton Integrations guarantees that recursive execution will not occur.

## Supported Providers

WIP

- [GitHub App](docs/providers/github.md)
- [Slack App](docs/providers/slack.md)
- AWS SNS
- GCP PubSub
- CloudEvents

## Setup

TBW
