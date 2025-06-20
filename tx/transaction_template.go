package tx

import "log/slog"

type TransactionCallback func() (any, error)

type TransactionTemplate struct {
	logger    *slog.Logger
	txManager *TransactionManager
}

func NewTransactionTemplate(txManager *TransactionManager) *TransactionTemplate {
	return &TransactionTemplate{
		logger:    slog.With().WithGroup("TransactionTemplate"),
		txManager: txManager,
	}
}

func (tpl *TransactionTemplate) Execute(action TransactionCallback) (result any, err error) {
	tx := tpl.txManager.GetTransaction()
	result, err = action()
	if err != nil {
		errRollback := tpl.txManager.Rollback(tx)
		if errRollback != nil {
			err = errRollback
			tpl.logger.Warn("Error when current transaction is rolling back", slog.Any("error", errRollback))
		}
	} else {
		errCommit := tpl.txManager.Commit(tx)
		if errCommit != nil {
			err = errCommit
			tpl.logger.Warn("Error when current transaction is committing", slog.Any("error", errCommit))
		}
	}
	return
}
