package mysql

import (
	"errors"
	"os"

	"github.com/nusiss-capstone-project/campaign-center-api/server/log"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() (*gorm.DB, error) {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		return nil, errors.New("MYSQL_DSN is not set")
	}
	database, err := gorm.Open(gormmysql.Open(dsn), &gorm.Config{PrepareStmt: true, SkipDefaultTransaction: true})
	if err != nil {
		return nil, err
	}
	DB = database
	if err := registerWriteLoggingCallbacks(DB); err != nil {
		return DB, err
	}
	if err := ensureCampaignParticipantsSchema(DB); err != nil {
		return DB, err
	}
	if err := DB.AutoMigrate(
		&model.Campaign{},
		&model.CampaignLandingPage{},
		&model.CampaignLandingPageTranslation{},
		&model.AuditLog{},
		&model.CampaignPerformanceDaily{},
	); err != nil {
		return DB, err
	}
	return DB, nil
}

// ensureCampaignParticipantsSchema migrates the join-only participant table.
// AutoMigrate cannot safely ADD NOT NULL datetime columns onto legacy rows under
// MySQL strict mode, so we rebuild when the old wide schema is still present.
func ensureCampaignParticipantsSchema(db *gorm.DB) error {
	if !db.Migrator().HasTable("campaign_participants") {
		return db.AutoMigrate(&model.CampaignParticipant{})
	}

	// Previous mistaken migration used join_at; rename back to joined_at.
	hasJoinAt, err := hasColumn(db, "campaign_participants", "join_at")
	if err != nil {
		return err
	}
	hasJoinedAt, err := hasColumn(db, "campaign_participants", "joined_at")
	if err != nil {
		return err
	}
	if hasJoinAt && !hasJoinedAt {
		if err := db.Exec("ALTER TABLE campaign_participants CHANGE COLUMN join_at joined_at DATETIME(3) NOT NULL").Error; err != nil {
			return err
		}
	}

	needsRebuild, err := participantTableNeedsRebuild(db)
	if err != nil {
		return err
	}
	if !needsRebuild {
		return db.AutoMigrate(&model.CampaignParticipant{})
	}

	log.Logger.Infow("migrating campaign_participants to join-only schema")
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS campaign_participants_new (
			id BIGINT NOT NULL AUTO_INCREMENT,
			campaign_id BIGINT NOT NULL,
			user_id BIGINT NOT NULL,
			joined_at DATETIME(3) NOT NULL,
			created_at DATETIME(3) NOT NULL,
			updated_at DATETIME(3) NOT NULL,
			PRIMARY KEY (id),
			UNIQUE KEY uk_participant_campaign_user (campaign_id, user_id),
			KEY idx_participant_user (user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`INSERT IGNORE INTO campaign_participants_new (id, campaign_id, user_id, joined_at, created_at, updated_at)
		 SELECT
			id,
			campaign_id,
			user_id,
			COALESCE(NULLIF(joined_at, '0000-00-00 00:00:00'), NULLIF(updated_at, '0000-00-00 00:00:00'), CURRENT_TIMESTAMP(3)),
			COALESCE(NULLIF(joined_at, '0000-00-00 00:00:00'), NULLIF(updated_at, '0000-00-00 00:00:00'), CURRENT_TIMESTAMP(3)),
			COALESCE(NULLIF(updated_at, '0000-00-00 00:00:00'), CURRENT_TIMESTAMP(3))
		 FROM campaign_participants
		 WHERE campaign_id IS NOT NULL AND user_id IS NOT NULL`,
		`DROP TABLE campaign_participants`,
		`RENAME TABLE campaign_participants_new TO campaign_participants`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			return err
		}
	}
	return nil
}

func participantTableNeedsRebuild(db *gorm.DB) (bool, error) {
	hasOldCol, err := hasColumn(db, "campaign_participants", "join_status")
	if err != nil {
		return false, err
	}
	if hasOldCol {
		return true, nil
	}
	hasJoinedAt, err := hasColumn(db, "campaign_participants", "joined_at")
	if err != nil {
		return false, err
	}
	return !hasJoinedAt, nil
}

func hasColumn(db *gorm.DB, table, column string) (bool, error) {
	var n int64
	err := db.Raw(`
		SELECT COUNT(*) FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?`,
		table, column,
	).Scan(&n).Error
	return n > 0, err
}
