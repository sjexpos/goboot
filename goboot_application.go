package goboot

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/mbndr/figlet4go"
	goboot_fx "github.com/sjexpos/goboot/fx"
	"github.com/sjexpos/goboot/log"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

//go:embed default.yaml
var resources embed.FS

const libraryName = "Go-Boot"
const libraryVersion = "0.1.0"
const banner_default_text string = "Banner"
const application_banner_property_name = "application.banner"
const application_log_property_name = "application.log"
const application_name_property_name = "application.name"

func Run(fxOpts ...fx.Option) {
	// Initialize the application
	app, err := NewGobootApplication(fxOpts...)
	if err != nil {
		panic(err) // Handle error appropriately, e.g., log it or return it
	}

	// Run the application
	app.Run()
}

type GobootApplication struct {
	// Add fields as necessary for your application
	fxOpts []fx.Option
}

func NewGobootApplication(fxOpts ...fx.Option) (*GobootApplication, error) {
	return &GobootApplication{
		fxOpts: fxOpts,
	}, nil
}

func (app *GobootApplication) Run() {
	start := time.Now()
	log.MDC.Set(log.GO_ROUTINE_NAME_FIELD_NAME, "main")
	environment := app.prepareEnvironment()
	app.printBanner(environment)
	app.setupLogger(environment)
	wd, _ := os.Getwd()
	logger := slog.With()
	logger.Info(fmt.Sprintf("Starting Bootstrap using %v with PID %v (%v)", runtime.Version(), os.Getpid(), wd))
	environmentModule := app.createEnvironmentModule(environment)
	options := []fx.Option{
		fx.WithLogger(func() fxevent.Logger {
			return &goboot_fx.SlogLogger{Logger: logger}
		}),
		// fx.WithLogger(func() fxevent.Logger {
		// 	return &fxevent.NopLogger
		// }),
	}
	options = append(options, environmentModule)
	options = append(options, app.fxOpts...)
	options = append(options, fx.Invoke(func() {
		slog.Info(fmt.Sprintf("Completed initialization in %v", time.Since(start)))
	}))
	fx.New(options...).Run()
}

func (app *GobootApplication) prepareEnvironment() *viper.Viper {
	v := viper.New()
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // this is useful e.g. want to use . in Get() calls, but environmental variables to use _ delimiters (e.g. app.port -> APP_PORT)
	v.SetConfigType("yaml")
	data, errData := resources.ReadFile("default.yaml")
	if errData == nil {
		errMerge := v.MergeConfig(bytes.NewReader(data))
		if errMerge != nil {
			slog.Warn("Embed default.yaml was not successfully read, %s", errMerge)
		}
	} else {
		slog.Warn("Embed default.yaml was not found")
	}
	_, errCfgFile := os.Stat("./application.yaml")
	if errCfgFile == nil {
		v.SetConfigFile("./application.yaml")
		errMerge := v.MergeInConfig()
		if errMerge != nil {
			slog.Warn("application.yaml was not successfully read, %s", errMerge)
		}
	} else {
		slog.Debug("application.yaml was not found")
	}
	return v
}

func (app *GobootApplication) printBanner(v *viper.Viper) {
	bannerText := banner_default_text
	if v.IsSet(application_banner_property_name) {
		bannerText = v.GetString(application_banner_property_name)
	}
	ascii := figlet4go.NewAsciiRender()
	ascii.LoadFont("./")
	options := figlet4go.NewRenderOptions()
	options.FontName = "standard"
	renderStr, _ := ascii.RenderOpts(bannerText, options)
	fmt.Print(renderStr)
	fmt.Printf("  :: %v ::        (%v)\n", libraryName, libraryVersion)
	fmt.Println()
}

func (app *GobootApplication) setupLogger(v *viper.Viper) {
	strLogLevel := "INFO" // default log level
	if v.IsSet(application_log_property_name) {
		strLogLevel = v.GetString(application_log_property_name)
	}
	var level slog.Level
	level.UnmarshalText(([]byte)(strLogLevel))
	appName := libraryName
	if v.IsSet(application_name_property_name) {
		appName = v.GetString(application_name_property_name)
	}
	log.SetupRootLogger(appName, level)
}

func (app *GobootApplication) createEnvironmentModule(v *viper.Viper) fx.Option {
	annotations := app.createAnnotationsFromEnvironment(v)
	return fx.Module("env",
		fx.Provide(
			context.Background,
		),
		fx.Provide(annotations...),
	)
}

func (app *GobootApplication) createAnnotationsFromEnvironment(v *viper.Viper) []interface{} {
	annotations := make([]interface{}, 0)
	keys := v.AllKeys()
	for _, propertyName := range keys {
		propertyValue := v.Get(propertyName)
		//fmt.Printf("App Name: %s => %s\n", propertyName, propertyValue)
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
		} else if propertyType.Kind() == reflect.Bool {
			var boolValue = propertyValue.(bool)
			annotation := fx.Annotate(func() bool { return boolValue }, fx.ResultTags(`name:"`+propertyName+`"`))
			annotations = append(annotations, annotation)
		}
	}
	annotation := fx.Annotate(func() *viper.Viper { return v })
	annotations = append(annotations, annotation)
	return annotations
}
