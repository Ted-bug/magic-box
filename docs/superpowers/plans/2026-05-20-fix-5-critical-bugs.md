# 修复 5 个必须修复的 Bug 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 修复 optimization.md 中列出的 5 个高优先级 Bug，恢复代码正确性。

**Architecture:** 5 个 Bug 分布在 4 个文件中，彼此独立，可按任意顺序修复。每个 Bug 采用 TDD 方式：先写/补全测试验证 Bug 存在，再修复代码，最后确认测试通过。

**Tech Stack:** Go 1.25, logrus, gorm/gen

---

## File Structure

| 操作 | 文件路径 | 职责 |
|------|----------|------|
| 修改 | `cmd/gen-model/method/method.go` | 修复 IsEmpty() 逻辑 |
| 修改 | `pkg/logger/log.go` | 修复 SetFormatter 目标对象 + 子 Logger 丢失 skipCall |
| 修改 | `pkg/logger/log_test.go` | 修复 Mode 值 + 补充 Bug 验证测试 |
| 修改 | `cmd/gen-model/main.go` | 添加 flag.Parse() 调用 |

---

### Task 1: 修复 `CommonMethod.IsEmpty()` 逻辑反转

**Files:**
- Modify: `cmd/gen-model/method/method.go:9-13`

- [ ] **Step 1: 编写验证 IsEmpty 正确行为的测试**

在 `cmd/gen-model/method/method.go` 同目录下创建测试文件 `method_test.go`：

```go
package method

import "testing"

func TestIsEmpty(t *testing.T) {
	var m *CommonMethod
	if !m.IsEmpty() {
		t.Error("nil receiver 应该返回 true")
	}

	m = &CommonMethod{ID: 0}
	if !m.IsEmpty() {
		t.Error("ID 为 0 应该返回 true")
	}

	m = &CommonMethod{ID: 1}
	if m.IsEmpty() {
		t.Error("ID 非 0 应该返回 false")
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./cmd/gen-model/method/ -run TestIsEmpty -v`
Expected: FAIL — `nil receiver 应该返回 true` 和 `ID 为 0 应该返回 true` 断言失败

- [ ] **Step 3: 修复 IsEmpty 逻辑**

将 `cmd/gen-model/method/method.go:9-13` 从：

```go
func (m *CommonMethod) IsEmpty() bool {
	if m == nil || m.ID == 0 {
		return false
	}
	return true
}
```

改为：

```go
func (m *CommonMethod) IsEmpty() bool {
	return m == nil || m.ID == 0
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./cmd/gen-model/method/ -run TestIsEmpty -v`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add cmd/gen-model/method/method.go cmd/gen-model/method/method_test.go
git commit -m "fix: 修复 CommonMethod.IsEmpty() 返回值逻辑反转"
```

---

### Task 2: 修复 `NewLogger` 对全局 logger 设置格式而非自定义 logger

**Files:**
- Modify: `pkg/logger/log.go:51-53`

- [ ] **Step 1: 编写验证 SetFormatter 目标对象的测试**

在 `pkg/logger/log_test.go` 中添加测试：

```go
func TestNewLogger_TextFormatter(t *testing.T) {
	logger, err := NewLogger(Config{
		Level:  "info",
		Format: "text",
		Mode:   "command",
	})
	if err != nil {
		t.Fatalf("创建 logger 失败: %v", err)
	}
	_, ok := logger.Entry.Logger.Formatter.(*logrus.TextFormatter)
	if !ok {
		t.Error("非 JSON 模式下应使用 TextFormatter")
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./pkg/logger/ -run TestNewLogger_TextFormatter -v`
Expected: FAIL — `非 JSON 模式下应使用 TextFormatter` 断言失败，因为当前代码对全局 logger 设置了 TextFormatter，自定义 logger 仍使用默认格式

- [ ] **Step 3: 修复 SetFormatter 调用目标**

将 `pkg/logger/log.go:51-53` 从：

```go
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: time.RFC3339,
	})
```

改为：

```go
	logger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: time.RFC3339,
	})
```

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./pkg/logger/ -run TestNewLogger_TextFormatter -v`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add pkg/logger/log.go pkg/logger/log_test.go
git commit -m "fix: 修复 NewLogger 对全局 logger 设置格式而非自定义 logger 实例"
```

---

### Task 3: 修复 `NewLogger` 子 Logger 丢失 `skipCall` 字段

**Files:**
- Modify: `pkg/logger/log.go:86-91`

- [ ] **Step 1: 编写验证子 Logger skipCall 传递的测试**

在 `pkg/logger/log_test.go` 中添加测试：

```go
func TestNewLogger_SkipCall(t *testing.T) {
	logger, err := NewLogger(Config{
		Level:  "info",
		Format: "json",
		Mode:   "command",
	})
	if err != nil {
		t.Fatalf("创建 logger 失败: %v", err)
	}
	if logger.skipCall != 3 {
		t.Errorf("主 logger skipCall 应为 3，实际为 %d", logger.skipCall)
	}

	sub := logger.NewLogger("test-caller")
	if sub.skipCall != logger.skipCall {
		t.Errorf("子 logger skipCall 应为 %d，实际为 %d", logger.skipCall, sub.skipCall)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./pkg/logger/ -run TestNewLogger_SkipCall -v`
Expected: FAIL — `子 logger skipCall 应为 3，实际为 0` 断言失败

- [ ] **Step 3: 修复子 Logger 缺少 skipCall 赋值**

将 `pkg/logger/log.go:86-91` 从：

```go
func (l *Logger) NewLogger(call string) *Logger {
	entry := l.Entry.WithField("caller", call)
	return &Logger{
		Entry: entry,
	}
}
```

改为：

```go
func (l *Logger) NewLogger(call string) *Logger {
	entry := l.Entry.WithField("caller", call)
	return &Logger{
		Entry:    entry,
		skipCall: l.skipCall,
	}
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./pkg/logger/ -run TestNewLogger_SkipCall -v`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add pkg/logger/log.go pkg/logger/log_test.go
git commit -m "fix: 修复子 Logger 创建时丢失 skipCall 字段"
```

---

### Task 4: 修复测试文件中 Mode 值不匹配

**Files:**
- Modify: `pkg/logger/log_test.go:12`

- [ ] **Step 1: 修复 Mode 值**

将 `pkg/logger/log_test.go:12` 从：

```go
		Mode:   "console",
```

改为：

```go
		Mode:   "command",
```

- [ ] **Step 2: 运行全部 logger 测试确认通过**

Run: `go test ./pkg/logger/ -v`
Expected: PASS — 所有测试通过

- [ ] **Step 3: 提交**

```bash
git add pkg/logger/log_test.go
git commit -m "fix: 修复测试中 Mode 值从 console 改为 command 与代码逻辑一致"
```

---

### Task 5: 修复 `main.go` 中 `flag.Parse()` 未调用

**Files:**
- Modify: `cmd/gen-model/main.go:15-16`

- [ ] **Step 1: 添加 flag.Parse() 调用**

将 `cmd/gen-model/main.go:15-16` 从：

```go
func main() {
	conf := NewConfig(*config)
```

改为：

```go
func main() {
	flag.Parse()
	conf := NewConfig(*config)
```

- [ ] **Step 2: 确认编译通过**

Run: `go build ./cmd/gen-model/`
Expected: 编译成功，无错误

- [ ] **Step 3: 提交**

```bash
git add cmd/gen-model/main.go
git commit -m "fix: 添加 flag.Parse() 调用使命令行参数生效"
```

---

## Self-Review

**1. Spec coverage:** 5 个 Bug 全部覆盖——#1 IsEmpty 逻辑反转（Task 1）、#2 SetFormatter 目标错误（Task 2）、#3 子 Logger 丢失 skipCall（Task 3）、#4 测试 Mode 值不匹配（Task 4）、#5 flag.Parse 未调用（Task 5）。

**2. Placeholder scan:** 无 TBD、TODO、fill in details 等占位符。所有步骤均包含完整代码。

**3. Type consistency:** `skipCall` 字段类型为 `int`，测试和实现中一致使用 `int`。`Mode` 字段值为 `"command"`，与 `log.go` switch case 匹配。
