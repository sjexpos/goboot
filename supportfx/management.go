package supportfx

import (
	"context"
	"fmt"
	"github.com/sjexpos/goboot/management"
	"log/slog"
	"net"
	"net/http"

	"go.uber.org/fx"
)

var ManagementModule = fx.Module("management",
	fx.Provide(
		fx.Private,
		fx.Annotate(
			func(managementPort int) *http.Server {
				mux := http.NewServeMux()
				mux.Handle("/actuator/", management.NewActuators())
				return &http.Server{
					Addr:    fmt.Sprintf(":%v", managementPort),
					Handler: mux,
				}
			},
			fx.ParamTags(`name:"management.server.port"`),
			fx.OnStart(func(server *http.Server) error {
				ln, err := net.Listen("tcp", server.Addr)
				if err != nil {
					return err
				}
				go server.Serve(ln)
				return nil
			}),
			fx.OnStop(func(ctx context.Context, srv *http.Server) error {
				slog.Info("Shutting down Http server")
				return srv.Shutdown(ctx)
			}),
		),
	),
	fx.Invoke(
		func(server *http.Server) {
			slog.Info(fmt.Sprintf("Management server started on port %v (%v) with context path '%v'", server.Addr, "http", "/"))
		},
	),
)
