package types

import (
	"encoding/json"
	"fmt"
	"gorm.io/datatypes"
	"testing"
	"time"
)

func TestNullJson(t *testing.T) {
	type MyCustom struct {
		F1 string `json:"f1"`
		F2 int    `json:"f2"`
	}
	type MyStruct struct {
		ID        int64           `json:"id"`
		Age       int             `json:"age"`
		Hobby     Null[string]    `json:"hobby"`
		Create    Null[time.Time] `json:"create"`
		Price     Null[float64]   `json:"price"`
		MyCustom  Null[MyCustom]  `json:"my_custom"`
		MyCustom2 MyCustom        `json:"myCustom2"`
	}
	myStruct := MyStruct{
		ID:     1,
		Age:    18,
		Hobby:  NewNull[string]("football", true),
		Create: NewNull[time.Time](time.Now(), true),
		Price:  NewNull[float64](9.99, false),
	}
	fmt.Printf("%v\n", myStruct)
	jsonData, err := json.Marshal(myStruct)
	if err != nil {
		t.Errorf("Error marshaling JSON: %v", err)
	}
	t.Logf("JSON Data: %s", jsonData)
}

func TestGormType(t *testing.T) {
	type Like struct {
		Name string `json:"name"`
	}
	// datatypes.JSONType[]处理可为NULL的mysql类型
	// 测试结果：能自行序列化/反序列化到数据库；json时能正确处理NULL值显示
	// 缺点：取值Data()时需要手动类型转换；无法判断该类型 当前是否有值！！
	type MyStruct struct {
		ID    int                       `json:"id"`
		Hobby datatypes.JSONType[*Like] `json:"hobby"`
	}
	m := MyStruct{
		ID:    1,
		Hobby: datatypes.NewJSONType[*Like](&Like{Name: "test"}),
	}
	jsonData, err := json.Marshal(m)
	if err != nil {
		t.Errorf("Error marshaling JSON: %v", err)
	}
	t.Logf("JSON Data: %s", jsonData)

	m2 := MyStruct{}
	str := `{"id":1,"hobby":null}`
	if err := json.Unmarshal([]byte(str), &m2); err != nil {
		t.Errorf("Error unmarshaling JSON: %v", err)
	}
	like := m2.Hobby.Data()
	if like != nil && like.Name == "like" {
		t.Log(like.Name)
	}
	t.Logf("%#v", m2.Hobby.Data())

	// datatypes.NULL[T]能够处理NULL问题
	// 缺点：缺少json反/序列化的NULL显示处理
	// 优点：直接判断类型指针是否有效
}
