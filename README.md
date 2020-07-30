# ambassador-external-filter-userinfo
## Background
Starting point: https://github.com/datawire/ambassador-auth-service

Given a `GET /oauth2/v3/userinfo -H "Host: {{HOSTNAME}} -H "Authorization: Bearer {{JWT}}"` return either 200 (success) or 511 for Authentication failure or 502 for IdP request failure.

## Implementation details:
- Userinfo endpoint can be obtained by doing `GET /.well-known/openid-configuration Host: server` and searching for key `userinfo_endpoint` from JSON body

- Pass the header "Authorization-URL" for the OIDC discovery endpoint.  "/.well-known/openid-configuration" is appended to the end.  (i.e. curl https://localhost:8080 -H "Authorization-URL: https://login.microsoftonline.com/common/v2.0")

- "Authorization" header is received from initial request and passed to the userinfo endpoint for validation.