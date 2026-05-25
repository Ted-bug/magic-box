package types

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

/*
自定义struct，需要实现Scan/Value，以便让GORM处理
sql.Null功能：1 零值有效的场景；2 允许Null值的场景；3 用户操作后才从Null转换为确切值
	默认值：V为类性零值，Valid为false，gorm忽略，不生成SQL
	设置有效值：V为非类型零值，Valid为true，gorm生成 k = v 的SQL
	设置Null值：V为非类型零值，Valid为false，gorm生成 k = NULL的SQL
*/

func NewNull[T any](v T, valid bool) Null[T] {
	return Null[T]{
		Null: sql.Null[T]{
			Valid: valid,
			V:     v,
		},
	}
}

func NewNullValid[T any](v T) Null[T] {
	return NewNull(v, true)
}

func NewNullInvalid[T any](v T) Null[T] {
	return NewNull(v, false)
}

type Null[T any] struct {
	sql.Null[T]
}

// MarshalJSON 实现json序列化接口
func (n Null[T]) MarshalJSON() ([]byte, error) {
	// 如果值无效，返回null
	if !n.Valid {
		return []byte("null"), nil
	}

	// 获取值的类型
	v := reflect.ValueOf(n.V)
	t := v.Type()
	if t == reflect.TypeOf(time.Time{}) {
		// 检查时间是否为零值
		realV := any(n.V).(time.Time)
		if realV.IsZero() {
			return []byte("null"), nil
		}
		return json.Marshal(realV.Format("2006-01-02 15:04:05"))
	}

	// 对于基础类型，返回对应零值或实际值
	// 如果是自定义结构体，调用它实现的json方法
	return json.Marshal(n.V)
}

// UnmarshalJSON 实现json反序列化接口
func (n *Null[T]) UnmarshalJSON(data []byte) error {
	// 如果是null，设置为无效
	if string(data) == "null" {
		n.Valid = false
		var zero T
		n.V = zero
		return nil
	}

	// 解析JSON数据
	if err := json.Unmarshal(data, &n.V); err != nil {
		return err
	}

	n.Valid = true
	return nil
}

func (n Null[T]) String() string {
	if !n.Valid {
		return "nil"
	}

	// 获取值的类型
	v := reflect.ValueOf(n.V)
	t := v.Type()

	// 对于时间类型
	if t == reflect.TypeOf(time.Time{}) {
		// 检查时间是否为零值
		realV := any(n.V).(time.Time)
		if realV.IsZero() {
			return "nil"
		}
		return realV.Format(time.DateTime)
	}

	// 对于基础类型和自定义结构体
	return fmt.Sprintf("%v", n.V)
}

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
