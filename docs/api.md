# API specifications

Api endpoints reside under `/api`, with the endpoints themselves being versioned.

## Version 1

Version 1 of the API resides under `/api/v1`

### Data types

- Plugin:

  - `id`: `number` - The unique ID of the plugin
  - `name`: `string` - The name of the plugin
  - `summary_short`: `string` - A short description of the plugin
  - `summary_long`: `string` - A full description of the plugin
  - `current_version`: `string` - The latest version published for this plugin
  - `all_versions`: `[string]` - All versions this plugin has. Includes current one
  - `tags`: `[string]` - The tags asocciated with this plugin
  - `authod_id`: `number` - The user ID of author of this plugin
  - `type`: `string` - Type of the plugin. Valid values are `"plugin"` and `"widget"`

- NewPlugin:

  - `name`: `string` - The name of the plugin
  - `summary_short`: `string` - A short description of the plugin
  - `summary_long`: `string` - A full description of the plugin
  - `initial_version`: `string` - The latest version published for this plugin
  - `tags`: `[string]` - The tags asocciated with this plugin
  - `type`: `string` - Type of the plugin. Valid values are `"plugin"` and `"widget"`
  - `code`: `string` - The code of the first version
  - `aiscript_version`: `string` - The version of AIScript the first version targets

- UpdatePlugin:

  - `name`: `string | undefined` - The new name. Not required
  - `summary_short`: `string | undefined` - The new short description. Not required
  - `summary_long`: `string | undefined` - The new full description. Not required
  - `tags`: `[string] | undefined` - The new tags of the plugin. Not required
  - `type`: `string | undefined` - New type of the plugin. Valid values are `"plugin"` and `"widget"`. Not required

- PluginVersion:

  - `code`: `string` - The full code of this version
  - `aiscript_version`: `string` - The version of AIScript this plugin version is intended for

- NewVersion:
  - `code`: `string` - The full code of this version
  - `aiscript_version`: `string` - The version of AIScript this plugin is intended for
  - `version_name`: `string` - The name of the version

- IdValue:
  - `id`: `number` - Some user or plugin ID

- IdList:
  - `ids`: `[number]` - A list of IDs of users or plugins

- AuthState:
  - `state`: `number`
    - The next state of the authentication process.
    - Binary flag:
      - 0: Ok
      - 1: Fail
      - 2: Needs fido
      - 4: Needs totp
      - 8: Needs mail
  - `process_id_or_token`: `string` - The current process' id

- MfaData:
  - `value`: `string` - The value for the mfa type, usually the latest token
  - `process_id`: `string` - The id of the process for which this is
  - `type`: `number` - What type of mfa is being confirmed (same values as for `AuthState.state`)

- RegisterData:
  - `process_id`: `string` - The current process id
  - `value`: `string` - Whatever value is needed for the registration action

- Userdata:
  - `name`: `string` - The name of the account
  - `mail`: `string | null` - Mail of the account
  - `description`: `string` - Description of the account
  - `approved`: `bool` - Whether the account is approved
  - `user_admin`: `bool` - Whether the account can manage other accounts
  - `plugin_admin`: `bool` - Whether the account can manage plugins
  - `plugins_owned`: `[number]` - The IDs of the plugins the account owns
  - `links`: `[string]` - Links the account is associated with

### Endpoints

- /api/v1/plugins
  - GET:
    - A list of all plugins in json format
    - Receives: Nothing
    - Returns: Array of `Plugin`
  - POST:
    - Restricted: Create a new plugin
    - Receives: `NewPlugin`
    - Returns: Nothing
- /api/v1/plugins/{id}
  - GET:
    - Returns the plugin with the specified ID
    - Receives: Nothing
    - Returns `Plugin`
  - POST:
    - Restricted: Create a new version of the plugin
    - Receives: `NewVersion`
    - Returns: Nothing
  - PUT:
    - Restricted: Update a plugin with the specified ID
    - Receives `UpdatePlugin`
    - Returns: Nothing
  - DELETE:
    - Restricted: Delete a plugin
    - Receives: Nothing
    - Returns: Nothing
- /api/v1/plugins/{id}/{version}
  - GET:
    - Returns the specified version
    - Receives: Nothing
    - Returns: `PluginVersion`
  - DELETE:
    - Restricted: Delete a plugin version
    - Receives: Nothing
    - Returns Nothing

- /api/v1/login/start
  - GET:
    - Starts authentication for a login request
    - Receives: Username and password via http basic auth
    - Returns: `AuthState`, empty process id if `AuthState.state` is failure
- /api/v1/login/mfa
  - POST:
    - Continues a login request. Note: Mfa not implemented yet
    - Receives: `MfaData`
    - Returns: `AuthState`

- /api/v1/register/start
  - POST:
    - Start the registration process
    - Receives: `RegisterData`, `RegisterData.process_id` is irrelevant and can be empty
    - Returns: `RegisterData`, but `RegisterData.value` is empty
- /api/v1/register/password
  - POST:
    - Set a password for an active registration process
    - Receives: `RegisterData`, `RegisterData.value` is the raw password
    - Returns: `RegisterData`, but `RegisterData.value` is empty
- /api/v1/register/mail
  - POST:
    - Set the mail for an active registration process
    - Receives: `RegisterData`, `RegisterData.value` is the mail address
    - Returns: `RegisterData`, but `RegisterData.value` is empty
- /api/v1/register/description
  - POST:
    - Set the description for an active registration process
    - Receives: `RegisterData`, `RegisterData.value` is the description
    - Returns: `RegisterData`, but `RegisterData.value` is empty
- /api/v1/register/finalise
  - POST:
    - Complete an active registration process
    - Receives: `RegisterData`, `RegisterData.process_id` is irrelevant and can be empty
    - Returns: Nothing
- /api/v1/register/cancel
  - POST:
    - Cancel an active registration process
    - Receives: `RegisterData`, `RegisterData.process_id` is irrelevant and can be empty
    - Returns: Nothing

- /api/v1/admin/users/approve
  - POST:
    - (Restricted: Account admins only) Approve a user
    - Receives: `IdValue` containing ID of target user
    - Returns: Nothing
- /api/v1/admin/users/unapproved
  - GET:
    - (Restricted: Account admins only) Get all unapproved users
    - Receives: Nothing
    - Returns: `IdList`
- /api/v1/admin/users/promote-admin/plugins
  - POST:
    - (Restricted: Account admins only) Promote an account to plugin admin
    - Receives: `IdValue` containing ID of target user
    - Returns: Nothing
- /api/v1/admin/users/promote-admin/accounts
  - POST:
    - (Restricted: Account admins only) Promote an account to account admin
    - Receives: `IdValue` containing ID of target user
    - Returns: Nothing
- /api/v1/admin/users/userdata/{id}
  - GET:
    - (Restricted: Account admins only) Get the information about an account, including admin status
    - Receives: Nothing
    - Returns: `Userdata`

- /api/v1/admin/plugins/approve
  - POST:
    - (Restricted: Plugin admins only) Approve a plugin
    - Receives: `IdValue` containing ID of target plugin
    - Returns: Nothing
- /api/v1/admin/plugins/unapproved
  - GET:
    - (Restricted: Plugin admins only) Get all unapproved plugin
    - Receives: Nothing
    - Returns: `IdList`
