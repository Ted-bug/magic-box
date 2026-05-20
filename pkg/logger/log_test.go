package logger

import (
	"context"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestLog(t *testing.T) {
	logger, _ := NewLogger(Config{
		Level:  "info",
		Format: "json",
		Mode:   "command",
	})
	logger.Info(context.TODO(), "hello world")
}

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
