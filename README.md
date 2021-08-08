# Tekton Integrations

It provides functions to integrate Tekton with other systems (GitHub, Slack, etc).

- `Provider` presents credentials and settings for integrating with external systems.
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

## Setup

TBW

## Supported Providers

WIP

### GitHub App

[GitHub App](docs/providers/github.md) Provider

- Notification:
  - Set the Run status to the commit status.

### Slack App

[Slack App](docs/providers/slack.md) Provider

- Notification:
  - Post the status at the end of Run.
