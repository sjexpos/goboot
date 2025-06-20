package datasource

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"
)

func NewDatasource(host string, port int, username string, password string, dbname string, poolMaxIdleConnections int, poolMaxOpenConnections int, poolConnectionMaxLifetime time.Duration, poolConnectionMaxIdleTime time.Duration) *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, username, password, dbname)
	sqlDB, errDB := sql.Open("postgres", psqlInfo)
	if errDB != nil {
		slog.Error(errDB.Error())
		os.Exit(1)
	}
	sqlDB.SetMaxIdleConns(poolMaxIdleConnections)
	sqlDB.SetMaxOpenConns(poolMaxOpenConnections)
	sqlDB.SetConnMaxLifetime(poolConnectionMaxLifetime)
	sqlDB.SetConnMaxIdleTime(poolConnectionMaxIdleTime)
	_, sqlErr := sqlDB.Exec("SELECT 1")
	if sqlErr != nil {
		slog.Error(sqlErr.Error())
		os.Exit(1)
	}
	slog.Info("Database connection was set up")
	return sqlDB
}
