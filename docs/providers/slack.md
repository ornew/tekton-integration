# Slack App Integration

## Provider

```yaml
apiVersion: integrations.tekton.dev/v1alpha1
kind: Provider
metadata:
  name: slack-app
  namespace: default
spec:
  type: SlackApp
  slackApp:
    accessToken:
      secretRef:
        name: slack-app-credentials
        key: access-token
    channels:
      - name: general
        id: Cxxxxxx
```

## Features

- Notify the result of TaskRun/PipelineRun to Slack channels

