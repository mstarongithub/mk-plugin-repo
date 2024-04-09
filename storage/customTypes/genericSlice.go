package customtypes

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// A generic slice adjusted for being storable by sql
// Uses json to de- and encode
type GenericSlice[T any] []T

func (arr *GenericSlice[T]) Value() (driver.Value, error) {
	if arr == nil {
		return nil, nil
	}
	return json.Marshal(arr)
}

func (arr *GenericSlice[T]) Scan(value any) error {
	if value == nil {
		return nil
	}
	valueBytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to cast value to byte slice: %v", value)
	}
	tmp := GenericSlice[T]{}
	err := json.Unmarshal(valueBytes, &tmp)
	if err != nil {
		return fmt.Errorf("failed to unmarshal value %v: %w", valueBytes, err)
	}
	*arr = tmp
	return nil
}
