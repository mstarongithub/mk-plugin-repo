# API documentation

## Structure
1. Definitions
2. Endpoints

## Definitions

### Endpoint definition structure

Each endpoint will have the following structure:
```
- Endpoint name:
    - Url: Full Url of the endpoint. May contain path parameters identified via `{name}`
    - Method: HTTP Method used
    - Restrictions: What restrictions are applied when calling it. See 
    - In data: Example json object. May be empty if not required
    - Out data: Example json object. May be empty if not required
```

### Restricitions

Endpoints may have restrictions for who can use them. These may depend on context and may be combined.
If multiple are combined, it's either x or y, not x and y

The following restrictions exist:
- Open: Everyone can access the endpoint
- Authenticated: A valid access token must be supplied. The tokens received from passkey authentication are also valid
- Owner: Only the authenticated owner of a resource may access it
- Plugin admin: The authenticated actor must be marked as plugin admin
- Account admin: The authenticated actor must be marked as account admin

### Errors

If an endpoint encounters an error, it will complete with an error message and appropriate status code.
The error message is a json object with the following form:
```json
{
    "id": 0123,
    "message": "Some message"
}
```
The id will be a valid id for one of the possible error types, as described in the following list.

0. Bad Request: Some part of the request to the endpoint was malformed
1. Database failure: A query to the database failed
2. Data not found: The requested data wasn't found
3. Json marshalling failed: The server failed to marshal the response data into a json object
4. Not approved: The provided actor (from token or public access) isn't approved for performing that action
5. Can't extend into past: The given timestamp is in the past and can't be used
6. Already exists: A resource with an equal name already exists
