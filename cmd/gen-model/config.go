package main

import (
	"encoding/json"
	"fmt"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	"io"
	"log"
	"os"
)

type Config struct {
	Type string `json:"type"` // 连接类型
	DSN  string `json:"dsn"`  // 数据库连接

	// 输出w物料配置
	OutPath      string `json:"out_path"`       // model文件输出位置
	OutFile      string `json:"out_file"`       // 输出文件后缀：默认.gen.go
	OnlyModel    bool   `json:"only_model"`     // 是否只输出model内容，不生成其它CURD的内容
	WithUnitTest bool   `json:"with_unit_test"` // 是否生成单元测试：固定false

	// 输出内容配置
	ModelPkgPath      string   `json:"model_pkg_path"`       // model包名
	WithImportPath    []string `json:"with_import_path"`     // 额外导入的pkg路径
	FieldNullable     bool     `json:"field_nullable"`       // db的null字段生成指针
	FieldCoverable    bool     `json:"field_coverable"`      // db存在默认零值，生成指针；gorm遇到类型零值不会参与sql操作
	FieldWithIndexTag bool     `json:"field_with_index_tag"` // db索引tag生成
	FieldWithTypeTag  bool     `json:"field_with_type_tag"`  // db列类型tag生成
	Tables            []string `json:"tables"`               // 需要生成的表
}

func NewConfig(file string) Config {
	confFile, err := os.Open(file)
	if err != nil {
		log.Fatalf("miss conf: %v", err)
	}
	confString, err := io.ReadAll(confFile)
	if err != nil {
		log.Fatalf("empty conf: %v", err)
	}
	if len(confString) == 0 {
		log.Fatalf("empty conf")
	}
	var conf Config
	if err := json.Unmarshal(confString, &conf); err != nil {
		log.Fatalf("read conf: %v", err)
	}
	return conf
}

// DataTypeMap 自定义类型映射
type DataTypeMap map[string]func(columnType gorm.ColumnType) (dataType string)

func NewDataTypeMap() DataTypeMap {
	m := make(DataTypeMap)
	typeMap := make(map[string][]string)
	typeMap["int"] = []string{"tinyint"}
	typeMap["int64"] = []string{"int", "bigint", "timestamp"}
	typeMap["float64"] = []string{"float", "double"}
	typeMap["string"] = []string{"tinytext", "text", "longtext", "json"}
	typeMap["time.Time"] = []string{"datetime", "date"}
	for goType, v := range typeMap {
		for _, t := range v {
			switch goType {
			case "int64", "int":
				m[t] = func(columnType gorm.ColumnType) (dataType string) {
					if columnType.Name() == "is_del" {
						return "soft_delete.DeletedAt"
					}
					if able, ok := columnType.Nullable(); ok && able {
						return fmt.Sprintf("common.NULL[%v]", goType)
					}
					return goType
				}
			case "string":
				m[t] = func(columnType gorm.ColumnType) (dataType string) {
					if able, ok := columnType.Nullable(); ok && able {
						return fmt.Sprintf("common.NULL[%v]", goType)
					}
					return goType
				}
			case "time.Time":
				m[t] = func(columnType gorm.ColumnType) (dataType string) {
					if able, ok := columnType.Nullable(); ok && able {
						return fmt.Sprintf("common.NULL[%v]", goType)
					}
					return goType
				}
			}
		}
	}
	return m
}

// FieldGormTagMap 追加字段的tag
type FieldGormTagMap map[string]func(tag field.GormTag) field.GormTag

func NewFieldTagMap() FieldGormTagMap {
	m := make(FieldGormTagMap)
	m["is_del"] = func(tag field.GormTag) field.GormTag {
		tag.Append("softDelete", "flag")
		return tag
	}
	return m
}
