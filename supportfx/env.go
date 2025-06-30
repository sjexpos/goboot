package supportfx

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"github.com/sjexpos/goboot/log"
	"log/slog"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mbndr/figlet4go"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

//go:embed *.yaml
var resources embed.FS

var EnvModule = fx.Module("env",
	fx.Provide(
		context.Background,
	),
	loadProperties(),
	fx.Invoke(
		fx.Annotate(
			func(app application) {
				if app.Banner == nil {
					text := BANNER_DEFAULT_TEXT
					app.Banner = &text
				}
				printBanner(*app.Banner)
				var level slog.Level
				level.UnmarshalText(([]byte)(app.Log))
				log.SetupRootLogger(app.Name, level)
				wd, _ := os.Getwd()
				slog.Info(fmt.Sprintf("Starting Bootstrap using %v with PID %v (%v)", runtime.Version(), os.Getpid(), wd))
			},
		),
	),
)

type application struct {
	Banner *string
	Name   string
	Log    string
}

const BANNER_DEFAULT_TEXT string = "Banner"

func loadProperties() fx.Option {
	annotations := make([]interface{}, 0)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // this is useful e.g. want to use . in Get() calls, but environmental variables to use _ delimiters (e.g. app.port -> APP_PORT)
	viper.SetConfigType("yaml")
	data, errData := resources.ReadFile("default.yaml")
	if errData == nil {
		errMerge := viper.MergeConfig(bytes.NewReader(data))
		if errMerge != nil {
			slog.Warn("Embed default.yaml was not successfully read, %s", errMerge)
		}
	} else {
		slog.Warn("Embed default.yaml was not found")
	}
	_, errCfgFile := os.Stat("./application.yaml")
	if errCfgFile == nil {
		viper.SetConfigFile("./application.yaml")
		errMerge := viper.MergeInConfig()
		if errMerge != nil {
			slog.Error("Error reading application file, %s", errMerge)
		}
	}
	keys := viper.AllKeys()
	for _, propertyName := range keys {
		propertyValue := viper.Get(propertyName)
		//mt.Printf("App Name: %s => %s\n", propertyName, propertyValue)
		propertyType := reflect.ValueOf(propertyValue)
		if propertyType.Kind() == reflect.String {
			durationValue, err := time.ParseDuration(propertyValue.(string))
			var annotation interface{}
			if err == nil {
				annotation = fx.Annotate(func() time.Duration { return durationValue }, fx.ResultTags(`name:"`+propertyName+`"`))
			} else {
				var stringValue string = propertyValue.(string)
				annotation = fx.Annotate(func() string { return stringValue }, fx.ResultTags(`name:"`+propertyName+`"`))
			}
			annotations = append(annotations, annotation)
		} else if propertyType.Kind() == reflect.Int {
			var intValue = propertyValue.(int)
			annotation := fx.Annotate(func() int { return intValue }, fx.ResultTags(`name:"`+propertyName+`"`))
			annotations = append(annotations, annotation)
		}
	}
	var app application
	err1 := viper.UnmarshalKey("application", &app)
	if err1 == nil {
		annotation := fx.Annotate(func() application { return app })
		annotations = append(annotations, annotation)
	}
	return fx.Module("settings", fx.Provide(annotations...))
}

func printBanner(text string) {
	ascii := figlet4go.NewAsciiRender()
	ascii.LoadFont("./")
	options := figlet4go.NewRenderOptions()
	options.FontName = "standard"
	renderStr, _ := ascii.RenderOpts(text, options)
	fmt.Print(renderStr)
	fmt.Printf("  :: Gin-Gonic ::        (%v)\n", gin.Version)
	fmt.Println()
}
