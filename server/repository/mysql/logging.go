package mysql

import (
	"time"

	"github.com/nusiss-capstone-project/campaign-center-api/server/log"
	"gorm.io/gorm"
)

// writeStartKey stores the write start time inside the GORM statement settings.
const writeStartKey = "cc:db_write_start"

// writeLoggedTables limits write logging to campaign admin tables.
var writeLoggedTables = map[string]struct{}{
	"campaigns":                    {},
	"campaign_drafts":              {},
	"campaign_reward_rules":        {},
	"campaign_user_rewards_ledger": {},
}

// registerWriteLoggingCallbacks logs every create/update/delete on campaign
// admin tables. Trace fields are attached from the statement context, so
// callers must run writes with db.WithContext(ctx).
func registerWriteLoggingCallbacks(db *gorm.DB) error {
	create := db.Callback().Create()
	update := db.Callback().Update()
	del := db.Callback().Delete()

	if err := create.Before("gorm:create").Register("cc:write_start_create", markWriteStart); err != nil {
		return err
	}
	if err := update.Before("gorm:update").Register("cc:write_start_update", markWriteStart); err != nil {
		return err
	}
	if err := del.Before("gorm:delete").Register("cc:write_start_delete", markWriteStart); err != nil {
		return err
	}
	if err := create.After("gorm:create").Register("cc:write_log_create", logWrite("create")); err != nil {
		return err
	}
	if err := update.After("gorm:update").Register("cc:write_log_update", logWrite("update")); err != nil {
		return err
	}
	if err := del.After("gorm:delete").Register("cc:write_log_delete", logWrite("delete")); err != nil {
		return err
	}
	return nil
}

func markWriteStart(db *gorm.DB) {
	db.Set(writeStartKey, time.Now())
}

func logWrite(op string) func(*gorm.DB) {
	return func(db *gorm.DB) {
		if db.Statement == nil {
			return
		}
		table := db.Statement.Table
		if _, ok := writeLoggedTables[table]; !ok {
			return
		}

		durationMs := 0.0
		if v, ok := db.Get(writeStartKey); ok {
			if start, ok := v.(time.Time); ok {
				durationMs = float64(time.Since(start).Microseconds()) / 1000
			}
		}

		logger := log.WithContext(db.Statement.Context)
		fields := []any{
			"db_op", op,
			"table", table,
			"rows_affected", db.Statement.RowsAffected,
			"duration_ms", durationMs,
		}
		if db.Error != nil {
			logger.Errorw("db_write_failed", append(fields, "error", db.Error.Error())...)
			return
		}
		logger.Infow("db_write", fields...)
	}
}
