package supportfx

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/sjexpos/goboot/core"
	goboot_gorm "github.com/sjexpos/goboot/gorm"
	"github.com/sjexpos/goboot/openapiv3"
	"github.com/sjexpos/goboot/swaggerui"
	"github.com/sjexpos/goboot/web"
	"github.com/spf13/viper"
	"github.com/wI2L/fizz"
	"github.com/wI2L/fizz/openapi"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

var WebModule = fx.Module("web",
	httpModule,
	ginModule,
)

var httpModule = fx.Module("http",
	fx.Provide(
		fx.Private,
		fx.Annotate(
			func(serverPort int, fizz *fizz.Fizz) (*http.Server, error) {
				return &http.Server{
					Addr:    fmt.Sprintf(":%v", serverPort),
					Handler: fizz,
				}, nil
			},
			fx.ParamTags(`name:"server.port"`),
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
			slog.Info(fmt.Sprintf("Http server started on port %v (%v) with context path '%v'", server.Addr, "http", "/"))
		},
	),
)

type gormParams struct {
	fx.In

	Middlewares              []web.Middleware `group:"gin-middlewares"`
	SwaggerUiPath            string           `name:"open-api-v3.swagger-ui.path"`
	ApiDocsPath              string           `name:"open-api-v3.api-docs.path"`
	OpenSessionInViewEnabled bool             `name:"gorm.open-session-in-view.enabled"`
	EM                       *gorm.DB         `optional:"true"`
}

var ginModule = fx.Module("gin",
	fx.Provide(
		gin.New,
		fizz.NewFromEngine,
	),
	fx.Decorate(
		fx.Annotate(
			func(gin *gin.Engine, params gormParams) (*gin.Engine, error) {
				if params.EM != nil && params.OpenSessionInViewEnabled {
					slog.Info("Open session in view enabled, adding OpenSessionInViewFilter")
					params.Middlewares = append(params.Middlewares, goboot_gorm.NewOpenSessionInViewFilter(params.EM))
				} else {
					slog.Warn("Open session in view disabled, not adding OpenSessionInViewFilter")
				}

				// Separa y ordena los middlewares que implementan Ordered
				type orderedMW struct {
					m web.Middleware
					o int
				}
				var ordered []orderedMW
				var unordered []web.Middleware

				for _, m := range params.Middlewares {
					if o, ok := any(m).(core.Ordered); ok {
						ordered = append(ordered, orderedMW{m, o.GetOrder()})
					} else {
						unordered = append(unordered, m)
					}
				}

				// Ordena los que implementan Ordered
				sort.Slice(ordered, func(i, j int) bool {
					return ordered[i].o < ordered[j].o
				})

				// Aplica primero los ordenados, luego los no ordenados
				for _, om := range ordered {
					gin.Use(om.m.DoFilter)
				}
				for _, m := range unordered {
					gin.Use(m.DoFilter)
				}
				swaggerui.Add(gin, params.SwaggerUiPath, params.ApiDocsPath)
				return gin, nil
			},
		),
	),
	fx.Decorate(
		fx.Annotate(
			registerOpenApi3Spec,
			fx.ParamTags(``, ``, `name:"server.port"`, `name:"open-api-v3.api-docs.path"`),
		),
	),
	fx.Invoke(
		func(fizz *fizz.Fizz) {
			slog.Info("Gin and Fizz were successfully configured")
		},
	),
)

const openApiV3PropertyNotFoundErrorMessage = "Property %s was not found, defaults will be used"
const openApiV3InfoPropertyName = "open-api-v3.info"
const openApiV3ServersPropertyName = "open-api-v3.servers"
const openApiV3SecurityRequirementPropertyName = "open-api-v3.securityRequirement"
const openApiV3SecuritySchemesPropertyName = "open-api-v3.securitySchemes"
const applicationNamePropertyName = "application.name"

func registerOpenApi3Spec(v *viper.Viper, fizz *fizz.Fizz, serverPort int, apiDocsPath string) (*fizz.Fizz, error) {
	var openInfo openapi.Info
	err1 := v.UnmarshalKey(openApiV3InfoPropertyName, &openInfo)
	if err1 != nil {
		slog.Debug(fmt.Sprintf(openApiV3PropertyNotFoundErrorMessage, openApiV3InfoPropertyName))
	}
	if len(openInfo.Title) == 0 && v.IsSet(applicationNamePropertyName) {
		openInfo.Title = v.GetString(applicationNamePropertyName)
	}
	var servers []*openapi.Server
	err1 = v.UnmarshalKey(openApiV3ServersPropertyName, &servers)
	if err1 != nil {
		slog.Debug(fmt.Sprintf(openApiV3PropertyNotFoundErrorMessage, openApiV3ServersPropertyName))
	}
	var securityRequirement []*openapi.SecurityRequirement
	err1 = v.UnmarshalKey(openApiV3SecurityRequirementPropertyName, &securityRequirement)
	if err1 != nil {
		slog.Debug(fmt.Sprintf(openApiV3PropertyNotFoundErrorMessage, openApiV3SecurityRequirementPropertyName))
	}
	var securitySchemes map[string]*openapi.SecuritySchemeOrRef
	err1 = v.UnmarshalKey(openApiV3SecuritySchemesPropertyName, &securitySchemes)
	if err1 != nil {
		slog.Debug(fmt.Sprintf(openApiV3PropertyNotFoundErrorMessage, openApiV3SecuritySchemesPropertyName))
	}
	openapiv3.RegisterOpenApi3Spec(fizz, &openInfo, servers, securityRequirement, securitySchemes, apiDocsPath, serverPort)
	return fizz, nil
}
