# API documentation

## Structure

1. Definitions
2. Endpoints

## Definitions

### Endpoint definition structure

Each endpoint will have the following structure:

```markdown
- Endpoint name:
  - Url: Full Url of the endpoint. May contain path parameters identified via `{name}`
  - Method: HTTP Method used
  - Description: A description of the endpoint's function
  - Restrictions: What restrictions are applied when calling it. See
  - In data: Example json object. May be empty if not required
  - Out data: Example json object. May be empty if not required
```

### Restrictions

Endpoints may have restrictions for who can use them.
These may depend on context and may be combined. If multiple are combined,
it's either x or y, not x and y

The following restrictions exist:

- Open: Everyone can access the endpoint
- Authenticated: A valid access token must be supplied.
  The tokens received from passkey authentication are also valid
- Owner: Only the authenticated owner of a resource may access it
- Plugin admin: The authenticated actor must be marked as plugin admin
- Account admin: The authenticated actor must be marked as account admin

### Errors

If an endpoint encounters an error, it will complete with an
error message and appropriate status code.
The error message is a json object with the following form:

```json
{
  "id": 0123,
  "message": "Some message"
}
```

The id will be a valid id for one of the possible error types,
as described in the following list.

0. Bad Request: Some part of the request to the endpoint was malformed
1. Database failure: A query to the database failed
2. Data not found: The requested data wasn't found
3. Json marshalling failed: The server failed to marshal the response data
   into a json object
4. Not approved: The provided actor (from token or public access) isn't approved
   for performing that action
5. Can't extend into past: The given timestamp is in the past and can't be used
6. Already exists: A resource with an equal name already exists

## Endpoints

### Plugins

- All plugins

  - Url: `/api/v1/plugins`
  - Method: `GET`
  - Description: Get a list of all (verified) plugins known
  - Restrictions: Open
  - Out Data:

    ```json
    [
      {
        "id": 1234, // uint: Id of the plugin
        "name": "Example plugin name", // string: Name of the plugin
        "summary_short": "Short description of the plugin", // string: Short description of the plugin
        "summary_long": "A fully detailed description of the plugin", // string: full description of a plugin, including details and configuration
        "current_version": "0.1.9a", // string: Latest version of the plugin
        "all_versions": ["0.1.9a", "0.1.8"], // array of strings: All versions of the plugin. Newest version first
        "tags": ["example", "alpha"], // array of strings: Tags the plugin fits with
        "author_id": 5678, // uint: Id of the author account
        "type": "plugin" // string: Type of the plugin. Must be either "plugin" or "widget"
      }
    ]
    ```

- Get specific plugin

  - Url: `/api/v1/plugins/{pluginId}`
  - Method: `GET`
  - Description: Get a specific plugin
  - Restrictions: Open
  - Out Data:

    ```json
    {
      "id": 1234, // uint: Id of the plugin
      "name": "Example plugin name", // string: Name of the plugin
      "summary_short": "Short description of the plugin", // string: Short description of the plugin
      "summary_long": "A fully detailed description of the plugin", // string: full description of a plugin, including details and configuration
      "current_version": "0.1.9a", // string: Latest version of the plugin
      "all_versions": ["0.1.9a", "0.1.8"], // array of strings: All versions of the plugin. Newest version first
      "tags": ["example", "alpha"], // array of strings: Tags the plugin fits with
      "author_id": 5678, // uint: Id of the author account
      "type": "plugin" // string: Type of the plugin. Must be either "plugin" or "widget"
    }
    ```

- New plugin

  - Url: `/api/v1/plugins`
  - Method: `POST`
  - Description: Create a new plugin
  - Restrictions: Authenticated
  - In Data:

    ```json
    {
      "name": "Plugin name", // string: Name of the new plugin
      "summary_short": "Short description", // string: Short description of the plugin
      "summary_long": "Long description", // string: Detailed description of the plugin
      "initial_version": "0.0.1a", // string: Name of the first version
      "code": "console.log(\"bob\")", // string: Code of the first version
      "tags": ["example", "alpha"], // array of strings: Tags the plugin falls under
      "type": "plugin", // string: Type of the plugin. Either "plugin" or "widget"
      "aiscript_version": "0.12.0" // string: Version of AI Script the first version is intended for
    }
    ```

- Update plugin

  - Url: `/api/v1/plugins/{pluginId}`
  - Method: `PUT`
  - Description: Update a targeted plugin
  - Restrictions: Owner + Plugin Admin
  - In Data:
    ```json
    {
      "name": "new name", // string: New name for the plugin. May be empty
      "summary_short": "new summary", // string: New short description for the plugin. May be empty
      "summary_long": "new description", // string: New detailed description for the plugin. May be empty
      "tags": ["new", "tags"], // array of strings: New tags the plugin falls under. May be empty
      "type": "widget" // string: New type of the plugin. Either "plugin" or "widget". May be empty
    }
    ```

- Delete plugin
  - Url: `/api/v1/plugins/{pluginId}`
  - Method: `DELETE`
  - Description: Delete a targeted plugin
  - Restrictions: Owner + Plugin

### Versions

- Get version

  - Url: `/api/v1/plugins/{pluginId}/{versionName}`
  - Method: `GET`
  - Description: Get a specific version of a plugin
  - Restrictions: Open
  - Out data:

    ```json
    {
      "code": "console.log('bob')", // string: Code of the plugin
      "aiscript_version": "0.12.0" // string: AI Script version the code was made for
    }
    ```

- New version

  - Url: `/api/v1/plugins/{pluginId}`
  - Method: `POST`
  - Description: Create a new version for a plugin
  - Restrictions: Owner
  - In data:

    ```json
    {
      "code": "console.log('new')", // string: Code of the new version
      "aiscript_version": "0.12.1", // string: AI Script version the new version is for
      "version_name": "0.1.9" // string: Name of the new version
    }
    ```

- Delete version
  - Url: `/api/v1/plugins/{pluginId}/{versionName}`
  - Method: `DELETE`
  - Description: Delete a version
  - Restrictions: Owner + Plugin Admin

### Tokens

- Get tokens

  - Url: `/api/v1/tokens`
  - Method: `GET`
  - Description: Get all tokens for the authenticated account
  - Restrictions: Authenticated
  - Out data:

    ```json
    {
      "tokens": [
        {
          "name": "gitlab", // string: Name of the token
          "token": "some random data", // string: The actual token
          "expires_at": "2009-11-10T23:00:00Z" // string: Timestamp when the token expires. RFC 3339, see https://pkg.go.dev/time#Time.MarshalJSON and https://pkg.go.dev/time#pkg-constants
        }
      ]
    }
    ```

- New token

  - Url: `/api/v1/tokens`
  - Method: `POST`
  - Description: Create a new token
  - Restrictions: Authenticated
  - In data:

    ```json
    {
      "name": "gitlab", // string: Name of the new token
      "expiration_date": "2009-11-10T23:00:00Z" // string: Timestamp when the token expires. RFC 3339, see https://pkg.go.dev/time#Time.MarshalJSON and https://pkg.go.dev/time#pkg-constants
    }
    ```

- Extend token

  - Url: `/api/v1/tokens/{name}`
  - Method: `PUT`
  - Description: Extend a token's expiry timestamp
  - Restrictions: Owner
  - In data:

    ```json
    {
      "extend_to": "2009-11-10T23:00:00Z" // string: New timestamp to where the token should be extended to. RFC 3339, see https://pkg.go.dev/time#Time.MarshalJSON and https://pkg.go.dev/time#pkg-constants
    }
    ```

- Delete token
  - Url: `/api/v1/tokens/{name}`
  - Method: `DELETE`
  - Description: Delete a token
  - Restrictions: Owner

### Account

- View account

  - Url: `/api/v1/accounts/{accountId}`
  - Method: `GET`
  - Description: View the public data of an account
  - Restrictions: Open
  - Out data:

    ```json
    {
      "name": "example name", // string: Name of the account
      "description": "Account description", // string: Description of the account
      "approved": true, // bool: Whether the account is approved
      "user_admin": false, // bool: Whether the account has user admin rights
      "plugin_admin": false, // bool: Whether the account has plugin admin rights
      "plugins_owned": [123, 456], // array of uints: List of plugin Ids the account owns
      "links": ["https://gitlab.com/examplename"] // array of strings: List of links the account wants to show
    }
    ```

- Delete account

  - Url: `/api/v1/delete`
  - Method: `POST`
  - Description: Delete an account
  - Restrictions: Owner + Account Admin
  - In data:

    ```json
    {
      "id": 123 // uint: Account Id to delete
    }
    ```

- Update account

  - Url: `/api/v1/accounts/update`
  - Method: `POST`
  - Description: Update an account's data
  - Restrictions: Owner + Account Admin
  - In data:

    ```json
    {
      "account_id": 123, // uint, optional: Id of the account to modify. If not same, requires account admin perms
      "name": "bob", // string, optional: New name of the account. Must not be taken by another account
      "description": "new description", string, optional: New description of the account
      "links": ["https://new.link"] // array of strings, optional: New list of links for the account
    }
    ```

### Admins

- Approve new account

  - Url: `/api/v1/admin/users/approve`
  - Method `POST`
  - Description: Approve a new account
  - Restrictions: Account Admin
  - In data:

    ```json
    {
      "id": 1234 // uint: Id of the account to approve
    }
    ```

- Get unapproved accounts

  - Url: `/api/v1/admin/users/unapproved`
  - Method: `GET`
  - Description: Get a list of account ids that haven't been approved yet
  - Restrictions: Account Admin
  - Out data:

    ```json
    {
      "accounts": [1234, 1235] // array of uints: List of account ids
    }
    ```

- Promote to plugin admin

  - Url: `/api/v1/admins/users/promote-admin/plugins`
  - Method: `POST`
  - Description: Promote an account to plugin admin
  - Restrictions: Account Admin
  - In data:

    ```json
    {
      "id": 1234 // uint: Id of account to promote
    }
    ```

- Promote to account admin

  - Url: `/api/v1/admins/users/promote-admin/accounts`
  - Method: `POST`
  - Description: Promote an account to account admin
  - Restrictions: Account Admin
  - In data:

    ```json
    {
      "id": 1234 // uint: Id of account to promote
    }
    ```

- Demote from plugin admin

  - Url: `/api/v1/admins/users/demote-admin/plugins`
  - Method: `POST`
  - Description: Demote an account from plugin admin.
    Note: Can't demote account with Id 1
  - Restrictions: Account Admin
  - In data:

    ```json
    {
      "id": 1234 // uint: Id of account to demote
    }
    ```

- Demote from account admin

  - Url: `/api/v1/admins/users/demote-admin/accounts`
  - Method: `POST`
  - Description: Demote an account from account admin.
    Note: Can't demote account with Id 1
  - Restrictions: Account Admin
  - In data:

    ```json
    {
      "id": 1234 // uint: Id of account to demote
    }
    ```

- Get user

  - Url: `/api/v1/admins/users/userdata/{id}`
  - Method: `GET`
  - Description: Get detailed data for a given account
  - Restrictions: Account Admin
  - Out data:

    ```json
    {
      "name": "bob", // string: Name of the account
      "mail": "bob@example.com", // string: Email of the account. May be null or empty
      "description": "bob is example", // string: Description of the account
      "approved": true, // bool: Whether the account has been approved
      "user_admin": false, // bool: Whether the account has user admin privileges
      "plugin_admin": false, // bool: Whether the account has plugin admin privileges
      "plugins_owned": [1234, 5678], // array of uints: List of plugin ids that the account owns
      "links": ["example.com"] // array of strings: List of urls the account owner can also be found
    }
    ```

- Approve plugin

  - Url: `/api/v1/admin/plugins/approve`
  - Method: `POST`
  - Description: Approve a new plugin
  - Restrictions: Plugin Admin
  - In data:

    ```json
    {
      "id": 1234 // uint: Id of the plugin to approve
    }
    ```

- Get unapproved plugins

  - Url: `/api/v1/admin/plugins/unapproved`
  - Method: `GET`
  - Description: Get all unapproved plugins
  - Restrictions: Plugin Admin
  - Out data:

    ```json
    {
      "plugins": [123, 124, 125] // array of uints: Ids of unapproved plugins
    }
    ```

## Missing endpoints

- Admin inspect unapproved plugin
