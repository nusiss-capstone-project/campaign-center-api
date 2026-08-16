package mysql

import (
	"database/sql"
	"os"
	"strconv"
	"time"

	"github.com/nusiss-capstone-project/campaign-center-api/server/log"
)

const (
	defaultMaxOpenConns         = 150
	defaultMaxIdleConns         = 10
	defaultConnMaxLifetime      = 30 * time.Minute
	defaultConnMaxIdleTime      = 5 * time.Minute
	defaultPoolStatsLogInterval = 30 * time.Second
)

func configureConnectionPool(sqlDB *sql.DB) {
	maxOpen := intEnv("MYSQL_MAX_OPEN_CONNS", defaultMaxOpenConns)
	maxIdle := intEnv("MYSQL_MAX_IDLE_CONNS", defaultMaxIdleConns)
	if maxIdle > maxOpen {
		maxIdle = maxOpen
	}
	lifetime := durationMinutesEnv("MYSQL_CONN_MAX_LIFETIME_MINUTES", defaultConnMaxLifetime)
	idleTime := durationMinutesEnv("MYSQL_CONN_MAX_IDLE_TIME_MINUTES", defaultConnMaxIdleTime)

	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxLifetime(lifetime)
	sqlDB.SetConnMaxIdleTime(idleTime)

	log.Logger.Infow("mysql_pool_config",
		"max_open_conns", maxOpen,
		"max_idle_conns", maxIdle,
		"conn_max_lifetime_minutes", lifetime.Minutes(),
		"conn_max_idle_time_minutes", idleTime.Minutes(),
	)
}

func startPoolStatsLogger(sqlDB *sql.DB) {
	interval := durationSecondsEnv("MYSQL_POOL_STATS_LOG_INTERVAL_SECONDS", defaultPoolStatsLogInterval)
	if interval <= 0 {
		return
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			stats := sqlDB.Stats()
			log.Logger.Infow("mysql_pool_stats", poolStatsFields(stats)...)
		}
	}()
}

func poolStatsFields(stats sql.DBStats) []any {
	return []any{
		"max_open_connections", stats.MaxOpenConnections,
		"open_connections", stats.OpenConnections,
		"in_use", stats.InUse,
		"idle", stats.Idle,
		"wait_count", stats.WaitCount,
		"wait_duration_ms", float64(stats.WaitDuration.Microseconds()) / 1000,
		"max_idle_closed", stats.MaxIdleClosed,
		"max_idle_time_closed", stats.MaxIdleTimeClosed,
		"max_lifetime_closed", stats.MaxLifetimeClosed,
	}
}

func intEnv(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		return fallback
	}
	return v
}

func durationMinutesEnv(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil || v < 0 {
		return fallback
	}
	return time.Duration(v * float64(time.Minute))
}

func durationSecondsEnv(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil || v < 0 {
		return fallback
	}
	return time.Duration(v * float64(time.Second))
}
