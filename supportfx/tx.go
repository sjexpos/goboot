package supportfx

import (
	"github.com/sjexpos/goboot/gorm"
	"github.com/sjexpos/goboot/tx"

	"go.uber.org/fx"
)

var TXModule = fx.Module("tx",
	fx.Provide(
		tx.NewTransactionManager,
		tx.NewTransactionTemplate,
		gorm.NewEntityManager,
	),
)
