package gorm

import (
	"github.com/sjexpos/goboot/tx"

	"gorm.io/gorm"
)

type EntityManager interface {
	Get() *gorm.DB
}

func NewEntityManager() EntityManager {
	return &entityManagerImpl{}
}

type entityManagerImpl struct {
}

func (e *entityManagerImpl) Get() *gorm.DB {
	return tx.GetTransactionSyncManager().GetResource()
}
