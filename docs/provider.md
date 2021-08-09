# Provider

The Provider provides the app's credentials and settings.
It can be referenced from multiple notifications.
Since each supported provider requires different parameters,
you should refer to each provider's documentation for detailed behavior.

Basic Example:

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

## Known Limits

## Status
