package supportfx

import (
	"goboot/gorm"
	"goboot/tx"

	"go.uber.org/fx"
)

var TXModule = fx.Module("tx",
	fx.Provide(
		tx.NewTransactionManager,
		tx.NewTransactionTemplate,
		gorm.NewEntityManager,
	),
)
