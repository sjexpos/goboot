package datasource

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"
)

func NewDatasource(host string, port int, username string, password string, dbname string, poolMaxIdleConnections int, poolMaxOpenConnections int, poolConnectionMaxLifetime time.Duration, poolConnectionMaxIdleTime time.Duration) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, username, password, dbname)
	sqlDB, errDB := sql.Open("postgres", psqlInfo)
	if errDB != nil {
		return nil, errDB
	}
	sqlDB.SetMaxIdleConns(poolMaxIdleConnections)
	sqlDB.SetMaxOpenConns(poolMaxOpenConnections)
	sqlDB.SetConnMaxLifetime(poolConnectionMaxLifetime)
	sqlDB.SetConnMaxIdleTime(poolConnectionMaxIdleTime)
	_, sqlErr := sqlDB.Exec("SELECT 1")
	if sqlErr != nil {
		return nil, sqlErr
	}
	slog.Info("Database connection was set up")
	return sqlDB, nil
}
