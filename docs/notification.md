# Notification

Notification provides a way to notify the provider of the status of the Run.
It also provides functions such as filtering notification targets and
suspending notifications. It doesn't touch how the notification is done.
The notification method depends on the implementation of the referenced provider.

Basic Example:

```yaml
apiVersion: integrations.tekton.dev/v1alpha1
kind: Notification
metadata:
  name: slack-notification
spec:
  suspend: false
  providerRef:
    name: slack-app
  filter:
    taskRun:
      enabled: false
    pipelineRun:
      enabled: true
    namespaceSelector:
      matchNames:
        - *
    labelSelector:
      matchLabels:
        foo: bar
      matchExpressions:
        - key: env
          operator: In
          values: [dev]
```

## Known Limits

## Status
