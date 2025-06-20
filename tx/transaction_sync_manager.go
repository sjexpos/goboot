package tx

import (
	"goboot/concurrent"

	"gorm.io/gorm"
)

var singleton *transactionSyncManager

func GetTransactionSyncManager() *transactionSyncManager {
	if singleton == nil {
		singleton = &transactionSyncManager{}
	}
	return singleton
}

type transactionSyncManager struct {
	resource concurrent.GoRoutineLocal[gorm.DB]
}

func (tsm *transactionSyncManager) BindResource(session *gorm.DB) {
	tsm.resource.Set(session)
}

func (tsm *transactionSyncManager) HasResource() bool {
	s := tsm.resource.Get()
	return s != nil
}

func (tsm *transactionSyncManager) UnbindResource() {
	tsm.resource.Clear()
}

func (tsm *transactionSyncManager) GetResource() *gorm.DB {
	return tsm.resource.Get()
}
