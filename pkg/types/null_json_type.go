package types

import (
	"database/sql/driver"
	"encoding/json"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type NullJSONType[T any] struct {
	Data  T
	Valid bool
}

func NewNullJSONType[T any](data T) NullJSONType[T] {
	return NullJSONType[T]{Data: data, Valid: true}
}

func NewNullJSONTypeNull[T any]() NullJSONType[T] {
	var zero T
	return NullJSONType[T]{Data: zero, Valid: false}
}

func (n NullJSONType[T]) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return json.Marshal(n.Data)
}

func (n *NullJSONType[T]) Scan(value interface{}) error {
	if value == nil {
		n.Valid = false
		var zero T
		n.Data = zero
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		str, ok := value.(string)
		if !ok {
			return json.Unmarshal(nil, &n.Data)
		}
		bytes = []byte(str)
	}

	n.Valid = true
	return json.Unmarshal(bytes, &n.Data)
}

func (n NullJSONType[T]) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.Data)
}

func (n *NullJSONType[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		var zero T
		n.Data = zero
		return nil
	}
	n.Valid = true
	return json.Unmarshal(data, &n.Data)
}

func (NullJSONType[T]) GormDataType() string {
	return "json"
}

func (NullJSONType[T]) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "JSON"
}
