package log

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lianjin/campaign-center-api/server/config"
	"github.com/lianjin/campaign-center-api/server/http/data"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger = zap.NewNop().Sugar()
)

func InitLogger() {
	writeSyncer := getLogWriter()
	encoder := getEncoder()
	core := zapcore.NewCore(encoder, writeSyncer, getLogLevel())
	if config.Config != nil && config.Config.LogConfig != nil && config.Config.LogConfig.FilePath != "" {
		consoleCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), getLogLevel())
		core = zapcore.NewTee(core, consoleCore)
	}
	Logger = zap.New(core, zap.AddCaller()).Sugar()
}

type ctxKey struct{}

func WithContext(ctx context.Context) *zap.SugaredLogger {
	if ctx == nil {
		return Logger
	}
	if l, ok := ctx.Value(ctxKey{}).(*zap.SugaredLogger); ok && l != nil {
		return l
	}
	return Logger
}

func TraceLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		span := trace.SpanFromContext(c.Request.Context())
		sc := span.SpanContext()

		logger := Logger
		if sc.IsValid() {
			logger = logger.With(
				"trace_id", sc.TraceID().String(),
				"span_id", sc.SpanID().String(),
				"service_name", data.ServiceName,
			)
		}

		ctx := context.WithValue(c.Request.Context(), ctxKey{}, logger)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func getLogLevel() zapcore.Level {
	level := zapcore.InfoLevel
	if config.Config != nil && config.Config.LogConfig != nil && config.Config.LogConfig.Level != "" {
		if err := level.UnmarshalText([]byte(config.Config.LogConfig.Level)); err != nil {
			level = zapcore.InfoLevel
		}
	}
	return level
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		location, err := time.LoadLocation("Asia/Singapore")
		if err != nil {
			location = time.Local
		}
		enc.AppendString(t.In(location).Format("2006-01-02 15:04:05"))
	}
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter() zapcore.WriteSyncer {
	if config.Config == nil || config.Config.LogConfig == nil || config.Config.LogConfig.FilePath == "" {
		return zapcore.AddSync(os.Stdout)
	}
	cwd, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("failed to get current working directory: %v", err))
	}
	logPath := filepath.Join(cwd, config.Config.LogConfig.FilePath)
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		panic(fmt.Sprintf("failed to create directories: %v", err))
	}
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		panic(err)
	}
	return zapcore.AddSync(file)
}
