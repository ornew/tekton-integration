
```yaml
apiVersion: integrations.tekton.dev/v1alpha1
kind: Notification
metadata:
  name: slack-notification
  namespace: default
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
