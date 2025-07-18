package gorm

import (
	"log/slog"

	"github.com/sjexpos/goboot/core"
	"github.com/sjexpos/goboot/tx"
	"github.com/sjexpos/goboot/web"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OpenSessionInViewFilter struct { // implements web.Middleware, core.Ordered
	logger *slog.Logger
	db     *gorm.DB
}

func (f *OpenSessionInViewFilter) DoFilter(c *gin.Context) {
	var session *gorm.DB
	var sessionCreated bool = false
	if !tx.GetTransactionSyncManager().HasResource() {
		f.logger.Debug("Opening Gorm Session in OpenSessionInViewFilter")
		session = f.db.Session(&gorm.Session{NewDB: true})
		tx.GetTransactionSyncManager().BindResource(session)
		sessionCreated = true
	} else {
		f.logger.Warn("Gorm Session was not created. It already exists one!")
	}
	defer func() {
		if sessionCreated {
			if tx.GetTransactionSyncManager().GetResource() == session {
				tx.GetTransactionSyncManager().UnbindResource()
				f.logger.Debug("Closing Gorm Session in OpenSessionInViewFilter")
			} else {
				f.logger.Warn("Gorm Session was created, but a different one is trying to close")
			}
		}
	}()
	c.Next()
}

func (*OpenSessionInViewFilter) GetOrder() int {
	return core.ORDERED_LOWEST_PRECEDENCE - 10
}

func NewOpenSessionInViewFilter(db *gorm.DB) web.Middleware {
	logger := slog.With().WithGroup("OpenSessionInViewFilter")
	return &OpenSessionInViewFilter{
		logger: logger,
		db:     db,
	}
}
