package tx

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

type TransactionManager struct {
	logger *slog.Logger
	db     *gorm.DB
}

func NewTransactionManager(db *gorm.DB) *TransactionManager {
	txManager := &TransactionManager{
		logger: slog.With().WithGroup("TransactionManager"),
		db:     db,
	}
	txManager.logger.Info("TransactionManager was created")
	return txManager
}

type transactionObject struct {
	parent *gorm.DB
}

func (txm *TransactionManager) GetTransaction() any {
	if GetTransactionSyncManager().HasResource() {
		local := GetTransactionSyncManager().GetResource()
		tx := local.Begin()
		txm.logger.Debug("New transaction was created from current session", slog.Any("session", fmt.Sprintf("%p", local)), slog.Any("tx", fmt.Sprintf("%p", tx)))

		txObject := &transactionObject{
			parent: local,
		}
		GetTransactionSyncManager().BindResource(tx)
		return txObject
	}
	tx := txm.db.Begin()
	txObject := &transactionObject{}
	txm.logger.Debug("New transaction was created from new session", slog.Any("session", fmt.Sprintf("%p", txm.db)), slog.Any("tx", fmt.Sprintf("%p", tx)))
	GetTransactionSyncManager().BindResource(tx)
	return txObject

}

func (txm *TransactionManager) Commit(status any) error {
	txObject := status.(*transactionObject)
	tx := GetTransactionSyncManager().GetResource()
	err := tx.Commit().Error
	if err == nil {
		txm.logger.Debug("Commit was called on transaction", slog.Any("tx", fmt.Sprintf("%p", tx)), slog.Any("session", fmt.Sprintf("%p", txObject.parent)))
	}
	GetTransactionSyncManager().UnbindResource()
	GetTransactionSyncManager().BindResource(txObject.parent)
	return err
}

func (txm *TransactionManager) Rollback(status any) error {
	txObject := status.(*transactionObject)
	tx := GetTransactionSyncManager().GetResource()
	err := tx.Rollback().Error
	if err == nil {
		txm.logger.Debug("Rollback was called on transaction", slog.Any("tx", fmt.Sprintf("%p", tx)), slog.Any("session", fmt.Sprintf("%p", txObject.parent)))
	}
	GetTransactionSyncManager().UnbindResource()
	GetTransactionSyncManager().BindResource(txObject.parent)
	return err
}
