# Get Authorization Info

GET https://api.webflow.com/v2/token/introspect

Information about the authorization tokenAccess to this endpoint requires a bearer token from either a [Data Client App](/data/docs/data-clients/getting-started), a [Workspace token](/data/reference/authentication/workspace-token), or a [Site token](/data/reference/authentication/site-token).The `application` object is only returned for tokens issued to an OAuth app. For tokens without an OAuth app, such as Site and Workspace API tokens (`grantType: client_credentials`), the response omits `application`.

Reference: https://developers.webflow.com/data/reference/token/introspect

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Response

### 200

Request was successful

- `authorization` (TokenIntrospectGetResponsesContentApplicationJsonSchemaAuthorization, optional) — The Authorization object
- `application` (TokenIntrospectGetResponsesContentApplicationJsonSchemaApplication, optional) — Details of the OAuth app the token belongs to. Present only for tokens issued to an OAuth app. Omitted for `client_credentials` tokens without an OAuth app, such as Site and Workspace API tokens.

## Types

### TokenIntrospectGetResponsesContentApplicationJsonSchemaAuthorization

The Authorization object

- `id` (string, optional) — The unique ID of the Authorization instance
- `createdOn` (string, optional) — The date the Authorization was created
- `lastUsed` (string, optional) — The date the Authorization was last used
- `grantType` (string, optional) — The grant type of the Authorization
- `rateLimit` (integer, optional) — The default rate limit for the Authorization (requests/min)
- `scope` (string, optional) — Comma separted list of OAuth scopes corresponding to the Authorization
- `authorizedTo` (TokenIntrospectGetResponsesContentApplicationJsonSchemaAuthorizationAuthorizedTo, optional)

### TokenIntrospectGetResponsesContentApplicationJsonSchemaApplication

Details of the OAuth app the token belongs to. Present only for tokens issued to an OAuth app. Omitted for `client_credentials` tokens without an OAuth app, such as Site and Workspace API tokens.

- `id` (string, optional) — Unique identifier for the Application
- `description` (string, optional) — Application description provided by the developer
- `homepage` (string, optional) — Application homepage URL provided by the developer
- `displayName` (string, optional) — Application name provided by the developer

### TokenIntrospectGetResponsesContentApplicationJsonSchemaDetailsItems

### TokenIntrospectGetResponses4XxContentApplicationJsonSchemaDetailsItems

### TokenIntrospectGetResponses5XxContentApplicationJsonSchemaDetailsItems

### TokenIntrospectGetResponsesContentApplicationJsonSchemaAuthorizationAuthorizedTo

- `siteIds` (list of any, optional) — Array of Sites this app is authorized to
- `workspaceIds` (list of any, optional) — Array of Workspaces this app is authorized to
- `userIds` (list of any, optional) — Array of Users this app is authorized to
