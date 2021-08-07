# GitHub App Integration

## Provider

```yaml
apiVersion: integrations.tekton.dev/v1alpha1
kind: Provider
metadata:
  name: github-app
  namespace: default
spec:
  type: GitHubApp
  githubApp:
    appId: 1
    privateKey:
      secretRef:
        name: github-app
        key: private-key.pem
```

## Features

- Sync TaskRun/PipelineRun Status to Commit Status
- Post the results of TaskRun/PipelineRun to PR

