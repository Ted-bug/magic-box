package logger

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

type Logger struct {
	*logrus.Entry

	mu          sync.Mutex // 保证顺序性
	lastTraceId string     // 多端同一次请求
	lastSpan    uint       // 单次请求的调用链

	skipCall int // 找到业务调用者所需的层级
}

type Config struct {
	SrvName  string `mapstructure:"srv_name"` // 服务名
	Level    string `mapstructure:"level"`    // trace, debug, info, warn, error, fatal, panic
	Format   string `mapstructure:"format"`   // text, json
	Mode     string `mapstructure:"output"`   // command,file
	Path     string `mapstructure:"path"`
	FileName string `mapstructure:"filename"`
	MaxAge   int    `mapstructure:"max_age"`   // 保留天数，单位天
	MaxSize  int    `mapstructure:"max_size"`  // 保留日志文件大小，单位MB
	SkipCall int    `mapstructure:"skip_call"` // 自定义的caller层级
	//MaxBackups uint   `mapstructure:"max_backups"` // 保留份数，单位个，暂时不启用，与max_age冲突
}

// NewLogger 根据配置创建一个新的Logger实例
func NewLogger(c Config) (*Logger, error) {
	logger := logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(c.Level)
	if err != nil {
		return nil, fmt.Errorf("无效的日志级别: %v", c.Level)
	}
	logger.SetLevel(level)

	// 设置日志输出格式
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: time.RFC3339,
	})
	if c.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	}

	// 设置日志输出位置
	var output io.Writer
	switch c.Mode {
	case "command":
		output = os.Stdout
	case "file":
		// 按日期切割
		fileName := fmt.Sprintf("%s/%s_output.", c.Path, c.FileName) + "%Y%m%d"
		output, err = rotatelogs.New(fileName,
			rotatelogs.WithMaxAge(time.Duration(c.MaxAge)*24*time.Hour),
			rotatelogs.WithRotationSize(int64(c.MaxSize)*1024*1024),
			//rotatelogs.WithRotationCount(c.MaxBackups), // MaxAge只能留一个
		)
		if err != nil {
			return nil, fmt.Errorf("初始化日志写入器错误：%v", err)
		}
	default:
		output = os.Stdout
	}
	logger.SetOutput(output)
	// TODO 自定义hook
	e := &Logger{Entry: logger.WithField("app", c.SrvName), skipCall: 3}

	return e, nil
}

func (l *Logger) NewLogger(call string) *Logger {
	entry := l.Entry.WithField("caller", call)
	return &Logger{
		Entry: entry,
	}
}

type KeyTraceKey struct{}

// SetTraceId 设置traceId
func SetTraceId(ctx context.Context) context.Context {
	return context.WithValue(ctx, KeyTraceKey{}, uuid.New().String())
}

// Trace 自动增加traceId
func (l *Logger) Trace(ctx context.Context) logrus.Fields {
	l.mu.Lock()
	defer l.mu.Unlock()
	nowTraceIdStr := ""
	if nowTraceId := ctx.Value(KeyTraceKey{}); nowTraceId != nil {
		if nowTraceId, ok := nowTraceId.(string); ok {
			nowTraceIdStr = nowTraceId
		}
	}
	if nowTraceIdStr == "" {
		nowTraceIdStr = uuid.New().String()
	}
	if nowTraceIdStr != l.lastTraceId {
		l.lastTraceId = nowTraceIdStr
		l.lastSpan = 0
	}
	l.lastSpan++
	return logrus.Fields{"trace_id": nowTraceIdStr, "span": l.lastSpan, "caller": getCaller(l.skipCall)}
}

// getCaller 获取调用者信息
func getCaller(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}
	// 提取文件名
	for i := len(file) - 1; i >= 0; i-- {
		if file[i] == '/' {
			file = file[i+1:]
			break
		}
	}
	return fmt.Sprintf("%v: %v", file, line)
}

func (l *Logger) Debug(ctx context.Context, args ...any) {
	l.Entry.WithFields(l.Trace(ctx)).Debug(args...)
}

func (l *Logger) Debugf(ctx context.Context, format string, args ...any) {
	l.Entry.WithFields(l.Trace(ctx)).Debugf(format, args...)
}

func (l *Logger) Info(ctx context.Context, args ...any) {
	l.Entry.WithFields(l.Trace(ctx)).Info(args...)
}

func (l *Logger) Infof(ctx context.Context, format string, args ...any) {
	l.Entry.WithFields(l.Trace(ctx)).Infof(format, args...)
}

func (l *Logger) Warn(ctx context.Context, args ...any) {
	l.Entry.WithFields(l.Trace(ctx)).Warn(args...)
}

func (l *Logger) Warnf(ctx context.Context, format string, args ...any) {
	l.Entry.WithFields(l.Trace(ctx)).Warnf(format, args...)
}

func (l *Logger) Error(ctx context.Context, args ...any) {
	l.Entry.WithFields(l.Trace(ctx)).Error(args...)
}

func (l *Logger) Errorf(ctx context.Context, format string, args ...any) {
	l.Entry.WithFields(l.Trace(ctx)).Errorf(format, args...)
}

func (l *Logger) Fatal(ctx context.Context, args ...any) {
	l.Entry.WithFields(l.Trace(ctx)).Fatal(args...)
}

func (l *Logger) Fatalf(ctx context.Context, format string, args ...any) {
	l.Entry.WithFields(l.Trace(ctx)).Fatalf(format, args...)
}

func (l *Logger) Panic(ctx context.Context, args ...any) {
	l.Entry.WithFields(l.Trace(ctx)).Panic(args...)
}

func (l *Logger) Panicf(ctx context.Context, format string, args ...any) {
	l.Entry.WithFields(l.Trace(ctx)).Panicf(format, args...)
}
