package gorm

import (
	"github.com/sjexpos/goboot/tx"
	"log/slog"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateOpenSessionInViewFilter(db *gorm.DB) gin.HandlerFunc {
	logger := slog.With().WithGroup("OpenSessionInViewFilter")
	return func(c *gin.Context) {
		var session *gorm.DB
		var sessionCreated bool = false
		if !tx.GetTransactionSyncManager().HasResource() {
			logger.Debug("Opening Gorm Session in OpenSessionInViewFilter")
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
					logger.Debug("Closing Gorm Session in OpenSessionInViewFilter")
				} else {
					logger.Warn("Gorm Session was created, but a different one is trying to close")
				}
			}
		}()
		c.Next()
	}
}
