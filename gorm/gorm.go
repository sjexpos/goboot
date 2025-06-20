package gorm

import (
	"database/sql"
	"log/slog"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func NewORM(sqlDB *sql.DB, logLevel string, slowThreshold time.Duration) *gorm.DB {
	gormDB, errGorm := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		QueryFields: true,
		//Logger:      logger.Default.LogMode(logger.Info),
		Logger:         NewSLog2(logLevel, slowThreshold),
		TranslateError: true,
	})
	if errGorm != nil {
		slog.Error(errGorm.Error())
		os.Exit(1)
	}
	slog.Info("Initialized Gorm for persistence")
	return gormDB
}
