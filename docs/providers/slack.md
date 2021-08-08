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
    channels:
      - name: general
      - id: C1234567890
```

## Features

- Notify the result of TaskRun/PipelineRun to Slack channels

## Setup

- Create a Slack app
- Add permissions: `chat:write`
- Install the app to your workspace and get bot user OAuth access token
- Create a Secret for Providers
- Create a Provider
- Create a Notification

Create a Slack app with `Create New App` on [Your Apps](https://api.slack.com/apps).

Open `OAuth & Permissions` tab, select `Add an OAuth Scope`.
You need to add `chat:write`.
This will authorize your app's bot to post to the added channel.
The bot can't post to channels that haven't invited, so invite bots if you need them.
Alternatively, you can give write permission to all public channels by `chat:write.public`.

Install the app to your workspace.
If you change the permissions after installation, you need to reinstall.
You get the Bot User OAuth Token here. This token is an access token that starts with `xoxb-`.

Create a Secret for Providers:

```sh
SECRET_NAME=slack-app

# required `private-key.pem`
kubectl create secret generic $SECRET_NAME --from-literal=access-key=xoxb-xxxx-xxxx-xxxx
```

will be created like:

```yaml
piVersion: v1
kind: Secret
metadata:
  name: slack-app
data:
  access-token: <base64-encoded-access-token>
```

Create a SlackApp Provider:

```yaml
apiVersion: integrations.tekton.ornew.io/v1alpha1
kind: Provider
metadata:
  name: slack-app
spec:
  type: SlackApp
  slackApp:
    channels:
      - name: general  # the channels what you want to notify.
    accessToken:
      secretRef:
        name: slack-app
```

Create a Notification:

```yaml
apiVersion: integrations.tekton.ornew.io/v1alpha1
kind: Notification
metadata:
  name: slack-app
spec:
  providerRef:
    name: slack-app
```

This will send a notification to Slack when the Pipeline Run finishes running.
