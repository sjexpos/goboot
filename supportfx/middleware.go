package supportfx

import (
	"github.com/sjexpos/goboot/web"
	"go.uber.org/fx"
)

func AddMiddleware(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(web.Middleware)),
		fx.ResultTags(`group:"gin-middlewares"`),
	)
}
