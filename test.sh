#!/bin/zsh

AAD_APP_ID="2c9e529a-3520-4bd7-9b6a-1b65749e6126"
AAD_SECRET="cQdPW~DY_k5NzRP5~cxC0RkD04S6dceYm."
AAD_TENANT_ID="ad3e52d4-a157-4546-822c-86716199e655"

curl -v -X POST -d "grant_type=client_credentials&client_id=$AAD_APP_ID&client_secret=$AAD_SECRET&resource=https%3A%2F%2Fmanagement.azure.com%2F" https://login.microsoftonline.com/$AAD_TENANT_ID/oauth2/token