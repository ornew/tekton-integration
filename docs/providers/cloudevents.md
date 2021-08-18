# CloudEvents

[CloudEvents](https://cloudevents.io/)

## Provider

### Webhook

Minimum example:

```yaml
apiVersion: integrations.tekton.dev/v1alpha1
kind: Provider
metadata:
  name: cloudevents
  namespace: default
spec:
  type: CloudEvents
  cloudEvents:
    protocol: WebHook
    webhook:
      url: https://xxxxx
```

Full example:

```yaml
apiVersion: integrations.tekton.dev/v1alpha1
kind: Provider
metadata:
  name: cloudevents
  namespace: default
spec:
  type: CloudEvents
  cloudEvents:
    protocol: WebHook
    # https://github.com/cloudevents/spec/blob/v1.0.1/http-webhook.md
    webhook:
      url: https://xxxxx
      authorization:
        # None (default), OAuth2, StaticToken
        type: OAuth2
        # We highly recommend using OAuth2.
        # Specify the token to be specified in the Authorization header.
        # You must also include the auth-schema (e.g. `Token my-access-token`).
        staticToken:
          secretRef:
            name: static-token
        # Set the access token obtained using OAuth2 in the Authorization header.
        # Supported client type is only `confidential`, not supported `public`.
        # https://datatracker.ietf.org/doc/html/rfc6749
        oauth2:
          # - (won't support) authorization_code ... https://datatracker.ietf.org/doc/html/rfc6749#section-4.1
          # - (won't support) token ... https://datatracker.ietf.org/doc/html/rfc6749#section-4.2
          # - (won't support) password ... https://datatracker.ietf.org/doc/html/rfc6749#section-4.3
          # - client_credentials ... https://datatracker.ietf.org/doc/html/rfc6749#section-4.4
          grantType: client_credentials
          endpoints:
            tokenURL: https://xxxxx/token
          # Scope is optional.
          scope: null
          # if grantType is client_credentials, this is required.
          # https://openid.net/specs/openid-connect-core-1_0.html#ClientAuthentication
          clientAuthentication:
            # - client_secret_basic         ... will be supported
            # - client_secret_post          ... will be supported
            # - client_secret_jwt           ... will be supported
            # - private_key_jwt             ... will be supported
            # - tls_client_auth             ... will be supported
            # - self_signed_tls_client_auth ... will be supported
            #
            #   method: client_secret_basic
            #
            # https://tools.ietf.org/html/rfc6749#section-2.3.1
            # https://tools.ietf.org/html/rfc7523#section-2.2
            #
            # If client_secret_* methods, required these values:
            #
            #   client-id: xxxxx
            #   client-secret: xxxxx
            #
            # If client_secret_jwt or private_key_jwt,
            # you need to specify the proper algorithm in response to your methods.
            # https://datatracker.ietf.org/doc/html/rfc7518#section-3.1
            #
            #   signing-algorithm: RS256
            #
            # If private_key_jwt, required the private key:
            #
            #   private-key.pem: ...
            #
            secretRef:
              name: cloudevents-oauth2-client-authentication-values
          additionalClaims:
            # The following claims will be set automatically:
            # - iss ... client-id from clientAuthentication
            # - sub ... client-id from clientAuthentication
            # - aud ... endpoints.tokenURL
            # - jti ... will be generated
            # - exp ... one hour later will be set
            # - iat ... will be set
            # You can set additional fixed parameters.
            # The above parameters cannot be overridden.
            # Do not include sensitive information.
            myclaim: foo
      # The validation request uses the HTTP OPTIONS method.
      # The request is directed to the exact resource target URI that is being registered.
      validation:
        # If enabled, perform a handshake with OPTIONS.
        enabled: false
        # If the handshake is successful, this value will be used in the Origin header.
        # If the WebHook-Allowed-Origin header is returned from the server, the validation passes.
        requestOrigin: my-cloudevent-origin
        # WebHook-Request-Callback is not supported.
        #requestCallback:
        # If not null and the WebHook-Allowed-Rate header is returned from the server, the validation passes.
        # The controller respects the responded rate.
        # However, the requests can be lost because rate limiting processes by the single controller.
        requestRate: 120
```
