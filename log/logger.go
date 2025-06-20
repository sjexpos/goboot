package log

import (
	"log/slog"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/rs/zerolog"
	slogzerolog "github.com/samber/slog-zerolog"
)

func SetupRootLogger(appName string, reportingLevel slog.Level) {
	var logger *slog.Logger

	// zapL := createZapLogger(zap.DebugLevel)
	// defer zapL.Sync()
	// logger = slog.New(zapslog.NewHandler(zapL.Core(), zapslog.WithCaller(true), zapslog.AddStacktraceAt(slog.LevelError)))

	zerologL := createZeroLogLogger(slogLevel2ZerologLevel(reportingLevel))
	logger = slog.New(NewSlogEnhancedHandler(appName, slogzerolog.Option{Logger: &zerologL}.NewZerologHandler(), "[ %6v]", "[ %12v]"))

	slog.SetDefault(logger)

}

func createZeroLogLogger(reportingLevel zerolog.Level) zerolog.Logger {
	zerolog.FormattedLevels = map[zerolog.Level]string{
		zerolog.TraceLevel: "TRACE",
		zerolog.DebugLevel: "DEBUG",
		zerolog.InfoLevel:  "INFO ",
		zerolog.WarnLevel:  "WARN ",
		zerolog.ErrorLevel: "ERROR",
		zerolog.FatalLevel: "FATAL",
		zerolog.PanicLevel: "PANIC",
	}
	zerolog.TimeFieldFormat = time.RFC3339Nano
	output := zerolog.ConsoleWriter{Out: os.Stdout}
	output.NoColor = false
	output.FormatTimestamp = ZerologConsoleFormatTimestamp(time.RFC3339Nano, nil, output.NoColor)
	output.PartsOrder = []string{
		zerolog.TimestampFieldName,
		zerolog.LevelFieldName,
		APP_ID_FIELD_NAME,
		//		enhancedlog.GO_ROUTINE_ID_FIELD_NAME,
		GO_ROUTINE_NAME_FIELD_NAME,
		zerolog.CallerFieldName,
		zerolog.MessageFieldName,
	}
	output.FieldsExclude = []string{
		APP_ID_FIELD_NAME,
		GO_ROUTINE_ID_FIELD_NAME,
		GO_ROUTINE_NAME_FIELD_NAME,
	}

	zerologL := zerolog.New(output).
		Level(reportingLevel).
		With().
		CallerWithSkipFrameCount(6).
		Timestamp().
		Logger()

	return zerologL
}

func slogLevel2ZerologLevel(level slog.Level) zerolog.Level {
	switch level {
	case slog.LevelDebug:
		return zerolog.TraceLevel
	case slog.LevelInfo:
		return zerolog.InfoLevel
	case slog.LevelWarn:
		return zerolog.WarnLevel
	case slog.LevelError:
		return zerolog.ErrorLevel
	default:
		return zerolog.PanicLevel
	}
}

func createZapLogger(reportingLevel zapcore.Level) *zap.Logger {
	encoderCfg := zap.NewDevelopmentEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("[")
		zapcore.CapitalColorLevelEncoder(l, enc)
		enc.AppendString("]")
	}
	encoderCfg.EncodeName = zapcore.FullNameEncoder
	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(reportingLevel),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          "console",
		EncoderConfig:     encoderCfg,
		OutputPaths: []string{
			"stderr",
		},
		ErrorOutputPaths: []string{
			"stderr",
		},
		InitialFields: map[string]interface{}{
			"pid": os.Getpid(),
		},
	}
	return zap.Must(config.Build())
}
