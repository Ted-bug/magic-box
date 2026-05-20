# Magic-Box 代码优化建议

## 一、Bug 修复（必须修复）

### 1. `CommonMethod.IsEmpty()` 逻辑反转

**文件：** `cmd/gen-model/method/method.go:9-13`

`IsEmpty()` 的返回值语义完全反了——当对象为空时返回 `false`，非空时返回 `true`。

```go
// 当前代码（错误）
func (m *CommonMethod) IsEmpty() bool {
    if m == nil || m.ID == 0 {
        return false  // 对象为空却返回 false
    }
    return true       // 对象非空却返回 true
}

// 修复后
func (m *CommonMethod) IsEmpty() bool {
    return m == nil || m.ID == 0
}
```

### 2. `NewLogger` 对全局 logger 设置格式而非自定义 logger

**文件：** `pkg/logger/log.go:51-53`

```go
// 当前代码（错误）—— 设置的是 logrus 全局 logger
logrus.SetFormatter(&logrus.TextFormatter{
    TimestampFormat: time.RFC3339,
})

// 修复后—— 应设置自定义的 logger 实例
logger.SetFormatter(&logrus.TextFormatter{
    TimestampFormat: time.RFC3339,
})
```

否则非 JSON 模式下，自定义 logger 不会使用 `TextFormatter`。

### 3. `NewLogger` 子 Logger 丢失 `skipCall` 字段

**文件：** `pkg/logger/log.go:86-91`

```go
// 当前代码（缺少 skipCall）
func (l *Logger) NewLogger(call string) *Logger {
    entry := l.Entry.WithField("caller", call)
    return &Logger{
        Entry: entry,
        // skipCall 未赋值，默认为 0，导致 getCaller 层级错误
    }
}

// 修复后
func (l *Logger) NewLogger(call string) *Logger {
    entry := l.Entry.WithField("caller", call)
    return &Logger{
        Entry:    entry,
        skipCall: l.skipCall,
    }
}
```

### 4. 测试文件中 Mode 值不匹配

**文件：** `pkg/logger/log_test.go:12`

测试中 `Mode: "console"`，但 `log.go` 的 switch 只匹配 `"command"`，实际走的是 `default` 分支。虽然结果恰好相同（都是 `os.Stdout`），但语义不对，应统一为 `"command"`。

### 5. `main.go` 中 `flag.Parse()` 未调用

**文件：** `cmd/gen-model/main.go:15`

定义了 `flag.String(...)` 并使用 `*config`，但从未调用 `flag.Parse()`，导致命令行参数永远不会生效。

```go
func main() {
    flag.Parse()  // 需要添加
    conf := NewConfig(*config)
    // ...
}
```

---

## 二、性能优化

### 6. `Logger.Trace()` 互斥锁导致并发瓶颈

**文件：** `pkg/logger/log.go:101-118`

每次日志调用都要加锁保护 `lastTraceId`/`lastSpan`，高并发场景下会成为瓶颈。而且多个 goroutine 共享同一个 Logger 实例的 span 计数器，会导致 span 编号混乱。

**建议方案：** 将 `traceId` 和 `span` 放到 `context` 中传递，使用 `atomic` 做 span 自增，无需 mutex：

```go
type traceCtx struct {
    traceID string
    span    uint64
}

func WithTrace(ctx context.Context) context.Context {
    return context.WithValue(ctx, KeyTraceKey{}, &traceCtx{
        traceID: uuid.New().String(),
    })
}

func (l *Logger) Trace(ctx context.Context) logrus.Fields {
    if v := ctx.Value(KeyTraceKey{}); v != nil {
        tc := v.(*traceCtx)
        span := atomic.AddUint64(&tc.span, 1)
        return logrus.Fields{
            "trace_id": tc.traceID,
            "span":     span,
            "caller":   getCaller(l.skipCall),
        }
    }
    return logrus.Fields{
        "trace_id": uuid.New().String(),
        "span":     1,
        "caller":   getCaller(l.skipCall),
    }
}
```

### 7. `Null[T]` 中使用反射判断 `time.Time`，性能较差

**文件：** `pkg/types/null_type.go:48-57` 和 `null_type.go:89-99`

`MarshalJSON` 和 `String()` 每次都通过 `reflect.ValueOf` 判断类型，性能较差。

**建议方案：** 使用泛型类型断言替代反射，速度快一个数量级：

```go
func (n Null[T]) MarshalJSON() ([]byte, error) {
    if !n.Valid {
        return []byte("null"), nil
    }
    switch v := any(n.V).(type) {
    case time.Time:
        if v.IsZero() {
            return []byte("null"), nil
        }
        return json.Marshal(v.Format("2006-01-02 15:04:05"))
    default:
        return json.Marshal(n.V)
    }
}

func (n Null[T]) String() string {
    if !n.Valid {
        return "nil"
    }
    switch v := any(n.V).(type) {
    case time.Time:
        if v.IsZero() {
            return "nil"
        }
        return v.Format(time.DateTime)
    default:
        return fmt.Sprintf("%v", n.V)
    }
}
```

---

## 三、代码质量

### 8. `NewDataTypeMap()` 中大量重复的闭包代码

**文件：** `cmd/gen-model/config.go:56-93`

`int64`、`string`、`time.Time` 三个分支的闭包逻辑几乎完全相同，可以提取为通用函数：

```go
func nullableTypeMapper(goType string) func(columnType gorm.ColumnType) string {
    return func(columnType gorm.ColumnType) string {
        if able, ok := columnType.Nullable(); ok && able {
            return fmt.Sprintf("common.NULL[%s]", goType)
        }
        return goType
    }
}
```

然后 `int64` 分支只需额外处理 `is_del` 的特殊情况即可。

### 9. `Config` 结构体中 `Type` 字段未使用

**文件：** `cmd/gen-model/config.go:14`

`Type string` 字段在代码中从未被读取或使用，属于死代码。`config.json` 中也没有这个字段，建议删除。

### 10. `Config` 中多个字段定义了但未使用

**文件：** `cmd/gen-model/config.go:27-29`

`FieldNullable`、`FieldCoverable`、`FieldWithIndexTag`、`FieldWithTypeTag` 这些字段在 `NewConfig` 中解析了，但在 `main.go` 中从未传给 `gen.Config`。如果确实需要这些配置，应该使用它们；否则应该删除。

### 11. `NewConfig` 中资源未关闭

**文件：** `cmd/gen-model/config.go:34`

`os.Open(file)` 打开了文件但从未 `defer confFile.Close()`，存在文件描述符泄漏。

```go
func NewConfig(file string) Config {
    confFile, err := os.Open(file)
    if err != nil {
        log.Fatalf("miss conf: %v", err)
    }
    defer confFile.Close()  // 需要添加
    // ...
}
```

### 12. `gen_test.go` 中测试函数为空

**文件：** `cmd/gen-model/gen_test.go:5-7`

`TestDefaultTypeMap` 是空函数，不会执行任何断言，应该补充测试逻辑或删除。

---

## 四、安全问题

### 13. `config.json` 中包含数据库密码

**文件：** `cmd/gen-model/config.json:1`

`root:root` 明文写了数据库用户名和密码。建议：

- 将 `config.json` 加入 `.gitignore`
- 提供一个 `config.example.json` 作为模板

---

## 五、架构建议

### 14. `Null[T]` 缺少 `Value()` 方法以完整支持 GORM

**文件：** `pkg/types/null_type.go`

`Null[T]` 嵌入了 `sql.Null[T]`，`sql.Null[T]` 已实现 `Scan/Value`，GORM 基本可用。但 `sql.Null[T]` 的 `Value()` 在 `Valid=false` 时返回 `nil`，在 `Valid=true` 且 `V` 为零值时也会返回零值——这可能不符合注释中描述的"零值不生成 SQL"的需求。如果需要 `FieldCoverable` 语义，需自行实现 `Value()` 方法。

---

## 优先级总结

| 优先级 | 编号 | 说明 |
|--------|------|------|
| 🔴 高 | #1, #2, #3, #5 | 逻辑错误，功能不正常 |
| 🟡 中 | #6, #7 | 高并发场景和频繁序列化场景 |
| 🟢 低 | #8, #9, #10, #11, #12 | 代码整洁度和可维护性 |
| 🔵 建议 | #13, #14 | 安全与最佳实践 |
