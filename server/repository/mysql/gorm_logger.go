package mysql

import (
	"context"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/nusiss-capstone-project/campaign-center-api/server/log"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

const defaultSlowSQLThreshold = 200 * time.Millisecond

// slowSQLLogger writes GORM slow SQL as structured single-line JSON logs.
type slowSQLLogger struct {
	level         gormlogger.LogLevel
	slowThreshold time.Duration
}

func newSlowSQLLogger() gormlogger.Interface {
	return &slowSQLLogger{
		level:         gormlogger.Warn,
		slowThreshold: slowSQLThreshold(),
	}
}

func (l *slowSQLLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	next := *l
	next.level = level
	return &next
}

func (l *slowSQLLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	if l.level >= gormlogger.Info {
		log.WithContext(ctx).Infow("gorm_info", "message", msg, "args", args)
	}
}

func (l *slowSQLLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	if l.level >= gormlogger.Warn {
		log.WithContext(ctx).Warnw("gorm_warn", "message", msg, "args", args)
	}
}

func (l *slowSQLLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	if l.level >= gormlogger.Error {
		log.WithContext(ctx).Errorw("gorm_error", "message", msg, "args", args)
	}
}

func (l *slowSQLLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.level <= gormlogger.Silent {
		return
	}
	elapsed := time.Since(begin)
	durationMs := float64(elapsed.Microseconds()) / 1000

	if err != nil && l.level >= gormlogger.Error && !errors.Is(err, gorm.ErrRecordNotFound) {
		sql, rows := fc()
		log.WithContext(ctx).Errorw("gorm_sql_error",
			"source", utils.FileWithLineNum(),
			"duration_ms", durationMs,
			"rows_affected", rows,
			"sql", sql,
			"error", err.Error(),
		)
		return
	}

	if elapsed >= l.slowThreshold && l.level >= gormlogger.Warn {
		sql, rows := fc()
		log.WithContext(ctx).Warnw("gorm_slow_sql",
			"source", utils.FileWithLineNum(),
			"duration_ms", durationMs,
			"threshold_ms", float64(l.slowThreshold.Microseconds())/1000,
			"rows_affected", rows,
			"sql", sql,
		)
	}
}

func slowSQLThreshold() time.Duration {
	raw := os.Getenv("GORM_SLOW_SQL_THRESHOLD_MS")
	if raw == "" {
		return defaultSlowSQLThreshold
	}
	ms, err := strconv.ParseFloat(raw, 64)
	if err != nil || ms <= 0 {
		return defaultSlowSQLThreshold
	}
	return time.Duration(ms * float64(time.Millisecond))
}
