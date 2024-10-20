package customtypes

import (
	"database/sql/driver"
	"fmt"
)

type PluginType int

const (
	PLUGIN_TYPE_PLUGIN = PluginType(iota)
	PLUGIN_TYPE_WIDGET
)

func (t *PluginType) Value() (driver.Value, error) {
	return t, nil
}

func (t *PluginType) Scan(value any) error {
	v, ok := value.(int64)
	if !ok {
		return fmt.Errorf("failed to cast value %+v as int. Type according to Go: %T", value, value)
	}
	*t = PluginType(v)
	return nil
}

func (t *PluginType) String() string {
	if *t == PLUGIN_TYPE_PLUGIN {
		return "Plugin"
	} else {
		return "Widget"
	}
}
