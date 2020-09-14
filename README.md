# ambassador-external-filter-userinfo
## Background:
Build an external filter that accepts the `Authorization` header from the previous filter in line and uses it to query the `User Info` endpoint.  This is done by doing a generic OIDC discovery for the endpoints, specifying the `User Info` endpoint and reapplying the `Authorization` header to the new request.  The `User Info` is then returned as a golang `string[map]`, which can be queried to generate response headers to send upstream.

## Installation Details:
This filter assumes you are either passing a valid `Authorization` header in the Request, or you have an upstream filter that gets the `Authorization` header for you.

1. Clone the repo: `git clone https://github.com/cakuros/ambassador-external-filter-userinfo`.
1. Run `docker build -t {{DOCKERHUB_REPO}}/{{IMAGE_NAME}}:{{VERSION}} .` (don't forget the "." at the end)
1. Push to repo with `docker push {{DOCKERHUB_REPO}}/{{IMAGE_NAME}}:{{VERSION}}`
1. Modify `k8s/deploy.yaml` in the environment variable `OIDC_SERVER` to point to the OIDC discovery point.  In the case of Azure AD, this is "https://login.microsoftonline.com/common/v2.0", the filter automatically appends "/.well-known/oidc-configuration/" for discovery.
1. Modify `k8s/deploy.yaml` to point to the image hosted on Dockerhub (caseykurosawa/external-filter-userinfo:1.5, for example)
1. Apply the deploy yaml and wait for it to spin up.
1. Change your `FilterPolicy` to include the new filter.
  ```yaml
  spec:
    rules:
    - filters:
      - name: filter1 # <-- OAuth filter that actually gets the Access Token
        arguments:
          scopes:
          - offline_access
      - name: external-filter-userinfo # <-- Custom Filter
      host: '*'
      path: /backend-debug/
  ```

## Making Changes
- in `main.go`, the User Info map is located at the end of the `handler` function, it also includes a sample `x-userinfo-name` header to add the value in `name` to a custom header.
- If you want to add more headers to send upstream, make sure you add them to the `Filter` under `spec.External.allowed_authorization_headers` in order for it to get passed along.


# Development notes (Not required for use)
## Background
Starting point: https://github.com/datawire/ambassador-auth-service

Given a `GET /oauth2/v3/userinfo -H "Host: {{HOSTNAME}} -H "Authorization: Bearer {{JWT}}"` return either 200 (success) or 511 for Authentication failure or 502 for IdP request failure.

## Implementation details:
- Userinfo endpoint can be obtained by doing `GET /.well-known/openid-configuration Host: server` and searching for key `userinfo_endpoint` from JSON body

- Pass the header "Authorization-URL" for the OIDC discovery endpoint.  "/.well-known/openid-configuration" is appended to the end.  (i.e. curl https://localhost:8080 -H "Authorization-URL: https://login.microsoftonline.com/common/v2.0")

- "Authorization" header is received from initial request and passed to the userinfo endpoint for validation.
