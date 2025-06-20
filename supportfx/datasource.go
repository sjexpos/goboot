package supportfx

import (
	"database/sql"
	"goboot/datasource"
	"log/slog"

	"go.uber.org/fx"
)

var DatasourceModule = fx.Module("datasource",
	fx.Provide(
		fx.Annotate(
			datasource.NewDatasource,
			fx.ParamTags(
				`name:"datasource.host"`,
				`name:"datasource.port"`,
				`name:"datasource.username"`,
				`name:"datasource.password"`,
				`name:"datasource.schema_name"`,
				`name:"datasource.pool.max_idle.connections"`,
				`name:"datasource.pool.max_open.connections"`,
				`name:"datasource.pool.max_lifetime.connection"`,
				`name:"datasource.pool.max_idle_time.connection"`,
			),
			fx.OnStop(func(ds *sql.DB) {
				slog.Info("Database shutdown")
				ds.Close()
			}),
		),
	),
)
