package main

import (
	"flag"
	"github.com/Ted-bug/magic-box/cmd/gen-model/method"
	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
	"log"
)

var config = flag.String("conf", "config.json", "生成所需的配置文件")

// 生成相关：https://gorm.io/gen/database_to_structs.html
func main() {
	flag.Parse()
	conf := NewConfig(*config)

	g := gen.NewGenerator(gen.Config{
		OutPath:      conf.OutPath,
		WithUnitTest: conf.WithUnitTest,
		ModelPkgPath: conf.ModelPkgPath,
	})
	db, err := gorm.Open(mysql.Open(conf.DSN))
	if err != nil {
		log.Fatalf("connect mysql err: %v", err)
	}
	g.UseDB(db)

	// model生成的相关配置
	g.WithDataTypeMap(NewDataTypeMap())
	g.WithImportPkgPath(conf.WithImportPath...)

	commonModelConf := []gen.ModelOpt{
		gen.WithMethod(method.CommonMethod{}),
		gen.FieldIgnore("create_time"),
		gen.FieldIgnore("update_time"),
	}
	for k, v := range NewFieldTagMap() {
		commonModelConf = append(commonModelConf, gen.FieldGORMTag(k, v))
	}
	for _, o := range commonModelConf {
		// 全局通用的model生成配置
		g.WithOpts(o)
	}

	var tables []any
	if len(conf.Tables) > 0 {
		for _, t := range conf.Tables {
			// 特定model的配置：g.GenerateModel(t, opt...)
			tables = append(tables, g.GenerateModel(t))
		}
	} else {
		tables = g.GenerateAllTable()
	}
	if !conf.OnlyModel {
		g.ApplyBasic(tables...)
	}
	g.Execute()
}
