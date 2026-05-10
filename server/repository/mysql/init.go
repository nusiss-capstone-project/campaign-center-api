package mysql

import (
	"fmt"
	"os"

	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func GetDSN() string {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME") // 这里读你新建的变量

	// 拼装符合 GORM 或 sql.DB 的 DSN
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, pass, host, port, dbname)
}

func Init() (*gorm.DB, error) {
	dsn := GetDSN()
	database, err := gorm.Open(gormmysql.Open(dsn), &gorm.Config{PrepareStmt: true, SkipDefaultTransaction: true})
	if err != nil {
		return nil, err
	}
	DB = database
	if err := DB.AutoMigrate(
		&model.Campaign{},
		&model.CampaignLandingPage{},
		&model.User{},
		&model.CampaignParticipant{},
		&model.RewardTransaction{},
		&model.AuditLog{},
	); err != nil {
		return DB, err
	}
	return DB, nil
}
