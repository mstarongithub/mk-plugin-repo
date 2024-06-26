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

### Endpoints

- /api/v1/plugins
  - GET:
    - A list of all plugins in json format
    - Receives: Nothing
    - Returns: Array of `Plugin`
  - POST:
    - (Restricted in the future) Create a new plugin
    - Receives: `NewPlugin`
    - Returns: Nothing
- /api/v1/plugins/{id}
  - GET:
    - Returns the plugin with the specified ID
    - Receives: Nothing
    - Returns `Plugin`
  - POST:
    - (Restricted in the future) Create a new version of the plugin
    - Receives: `NewVersion`
    - Returns: Nothing
  - PUT:
    - (Restricted in the future) Update a plugin with the specified ID
    - Receives `UpdatePlugin`
    - Returns: Nothing
  - DELETE:
    - (Restricted in the future) Delete a plugin
    - Receives: Nothing
    - Returns: Nothing
- /api/v1/plugins/{id}/{version}
  - GET:
    - Returns the specified version
    - Receives: Nothing
    - Returns: `PluginVersion`
  - DELETE:
    - (Restricted in the future) Delete a plugin version
    - Receives: Nothing
    - Returns Nothing
