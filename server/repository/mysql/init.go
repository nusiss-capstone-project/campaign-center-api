package mysql

import (
	"errors"
	"os"

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
	if err := DB.AutoMigrate(
		&model.Campaign{},
		&model.CampaignLandingPage{},
		&model.CampaignLandingPageTranslation{},
		&model.User{},
		&model.UserAuthMapping{},
		&model.CampaignParticipant{},
		&model.RewardTransaction{},
		&model.AuditLog{},
		&model.UserAccount{},
		&model.AccountTransaction{},
		&model.CampaignPerformanceDaily{},
	); err != nil {
		return DB, err
	}
	return DB, nil
}
