package repository

import (
	"fmt"

	"github.com/lianjin/campaign-center-api/server/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() (*gorm.DB, error) {
	if config.Config == nil || config.Config.MySQLConfig == nil || !config.Config.MySQLConfig.Enabled {
		DB = nil
		return nil, nil
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Config.MySQLConfig.UserName,
		config.Config.MySQLConfig.Password,
		config.Config.MySQLConfig.Host,
		config.Config.MySQLConfig.Port,
		config.Config.MySQLConfig.DBName,
	)
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{PrepareStmt: true, SkipDefaultTransaction: true})
	if err != nil {
		return nil, err
	}
	DB = database
	return DB, nil
}
