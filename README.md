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

- CRD References
  - [Provider](docs/provider.md)
  - [Notification](docs/notification.md)

## Supported Providers

WIP

- [GitHub App](docs/providers/github.md)
- [Slack App](docs/providers/slack.md)
- AWS SNS
- GCP PubSub
- CloudEvents

## Motivation

Developers today merge codes into baselines frequently, and CI systems need to
run automated builds, tests, and even deployments continuously, and keep
developers informed of the results.

There are multiple components in Tekton, with the main development of Pipelines,
Triggers and Dashboard. These do not support all in CI/CD systems and we will
actually need to work with other services. Although Tekton is designed as a
CI/CD system, it does not currently support first class integration with
external services.

The components that make up a practical CI are not only the pipeline engine
for automating workloads. It will has a code repository with version control
system that holds the shared mainline, a communication tool that enables
developers to quickly check results, a collaboration service for multiple
developers.

Tekton Integrations provide a common pattern for Tekton Pipelines to integrate
with other services, allowing you to quickly build CI/CD systems.

### Goal

- Notify the execution status of Tekton Pipeline to external services.

### Non-Goal (subject to change)

- Control the Tekton Pipeline from external services.
  - We are currently considering that this is likely to be provided by Triggers.
    However, there is room for consideration of the possibility that it will be
    provided as a higher level of functionality using Triggers, and that the
    pipeline may be controlled by a different route than Triggers.

## Requirements

1. Users can declaratively define how external services and pipelines work
   together.
2. Available in a format that follows Kubernetes conventions, so users do not
   have to pay an additional learning cost.

## Our Proposal

We define new CRDs that make up the integration.

One is `Provider` that provides settings and credentials for external services.
It can used Kubernetes role-based access controll.

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
```

The information required depends on the provider. Instead of adding resources
for each corresponding provider, it gives an abstraction as a provider.

The main way of integrations with external services the pipelines is
notification. We add a new CRD `Notification`.

```yaml
apiVersion: integrations.tekton.ornew.io/v1alpha1
kind: Notification
metadata:
  name: github-commit-statuses
spec:
  suspend: false
  providerRef:
    name: github-app
  filter:
    labelSelector:
      matchLabels:
        foo: bar
```

`Notification` do not know how notifications is done, but responsible to route
pipelines status to the provider following notification rules.

It can filters the pipeline status of interest by labels, namespace, kind,
status, values of results, etc., and send only the matched to the provider's
notification feature.

`Provider` can be referenced from multiple `Notification`.

How notifications are reflected in external services depends on the type of
provider. Due to service restrictions, all information may not be reflected.

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
If recursion occurs without proper resource quotas,
Run will increase exponentially and cause serious cluster-wide failure.
In my experience, the API server etcd went down and the cluster went out
of control for several hours. Tekton Integrations guarantees no recursive Run.

### Commit Status Tracker

- [tektoncd/experimental/commit-status-tracker](https://github.com/tektoncd/experimental/tree/main/commit-status-tracker)

It supports updating the commit status on GitHub. CRD is not added.
A single credential is supported for each controller.

### Tekton Notifiers

- [tektoncd/experimental/notifiers](https://github.com/tektoncd/experimental/tree/main/notifiers)

### CloudEvents Controller

- [tektoncd/community#435](https://github.com/tektoncd/community/issues/435)
- [tektoncd/experimental/cloudevents](https://github.com/tektoncd/experimental/tree/main/cloudevents)

### Tekton Results

- [tektoncd/results](https://github.com/tektoncd/results)

## How to Add a New Provider

TBW
