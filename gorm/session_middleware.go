package gorm

import (
	"errors"
	"goboot/tx"
	"log/slog"

	"gorm.io/gorm"
)

func CreateSessionMiddleware(db *gorm.DB) func(func()) {
	logger := slog.With().WithGroup("SessionMiddleware")
	return func(f func()) {
		var session *gorm.DB
		var sessionCreated bool = false
		if !tx.GetTransactionSyncManager().HasResource() {
			logger.Debug("Opening Gorm Session in SessionMiddleware")
			session = db.Session(&gorm.Session{NewDB: true})
			tx.GetTransactionSyncManager().BindResource(session)
			sessionCreated = true
		} else {
			logger.Warn("Gorm Session was not created. It already exists one!")
		}
		defer func() {
			if sessionCreated {
				if tx.GetTransactionSyncManager().GetResource() == session {
					tx.GetTransactionSyncManager().UnbindResource()
					logger.Debug("Closing Gorm Session in SessionMiddleware")
				} else {
					logger.Warn("Gorm Session was created, but a different one is trying to close")
				}
			}
		}()
		f()
	}
}

func ExpandMiddleware() func(func()) {
	if !tx.GetTransactionSyncManager().HasResource() {
		panic(errors.ErrUnsupported)
	}
	db := tx.GetTransactionSyncManager().GetResource()
	return CreateSessionMiddleware(db)
}
