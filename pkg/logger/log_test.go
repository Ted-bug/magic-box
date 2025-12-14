package logger

import (
	"context"
	"testing"
)

func TestLog(t *testing.T) {
	logger, _ := NewLogger(Config{
		Level:  "info",
		Format: "json",
		Mode:   "console",
	})
	logger.Info(context.TODO(), "hello world")
}
