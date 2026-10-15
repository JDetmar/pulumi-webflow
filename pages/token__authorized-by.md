# Get Authorization User Info

GET https://api.webflow.com/v2/token/authorized_by

Information about the Authorized User

Required Scope | `authorized_user:read`

Reference: https://developers.webflow.com/data/reference/token/authorized-by

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Response

### 200

Request was successful

- `id` (string, optional) — The unique ID of the user
- `email` (string, optional) — The user's email address
- `firstName` (string, optional) — The user's first name
- `lastName` (string, optional) — The user's last name

## Types

### TokenAuthorizedByGetResponsesContentApplicationJsonSchemaDetailsItems

### TokenAuthorizedByGetResponses4XxContentApplicationJsonSchemaDetailsItems

### TokenAuthorizedByGetResponses5XxContentApplicationJsonSchemaDetailsItems
