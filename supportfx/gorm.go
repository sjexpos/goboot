package supportfx

import (
	"goboot/gorm"

	"go.uber.org/fx"
)

var GormModule = fx.Module("gorm",
	fx.Provide(
		fx.Annotate(
			gorm.NewORM,
			fx.ParamTags(``, `name:"gorm.log.level"`, `name:"gorm.query.slow.threshold"`),
		),
	),
)
