package fx

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"go.uber.org/fx/fxevent"
)

// SlogLogger an Fx event logger that logs events using a slog logger.
type SlogLogger struct {
	Logger *slog.Logger

	ctx        context.Context
	logLevel   slog.Level
	errorLevel *slog.Level
}

// UseContext sets the context that will be used when logging to slog.
func (l *SlogLogger) UseContext(ctx context.Context) {
	l.ctx = ctx
}

// UseLogLevel sets the level of non-error logs emitted by Fx to level.
func (l *SlogLogger) UseLogLevel(level slog.Level) {
	l.logLevel = level
}

// UseErrorLevel sets the level of error logs emitted by Fx to level.
func (l *SlogLogger) UseErrorLevel(level slog.Level) {
	l.errorLevel = &level
}

func (l *SlogLogger) filter(fields []any) []any {
	filtered := []any{}

	for _, field := range fields {
		if field, ok := field.(slog.Attr); ok {
			if _, ok := field.Value.Any().(slogFieldSkip); ok {
				continue
			}
		}

		filtered = append(filtered, field)
	}

	return filtered
}

func (l *SlogLogger) logEvent(msg string, fields ...any) {
	l.Logger.Log(l.ctx, l.logLevel, msg, l.filter(fields)...)
}

func (l *SlogLogger) logDebugEvent(msg string, fields ...any) {
	l.Logger.DebugContext(l.ctx, msg, l.filter(fields)...)
}

func (l *SlogLogger) logInfoEvent(msg string, fields ...any) {
	l.Logger.InfoContext(l.ctx, msg, l.filter(fields)...)
}

func (l *SlogLogger) logError(msg string, fields ...any) {
	lvl := slog.LevelError
	if l.errorLevel != nil {
		lvl = *l.errorLevel
	}

	l.Logger.Log(l.ctx, lvl, msg, l.filter(fields)...)
}

// LogEvent logs the given event to the provided Zap logger.
func (l *SlogLogger) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		l.logDebugEvent("OnStart hook executing",
			slog.String("callee", e.FunctionName),
			slog.String("caller", e.CallerName),
		)
	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			l.logError("OnStart hook failed",
				slog.String("callee", e.FunctionName),
				slog.String("caller", e.CallerName),
				slogErr(e.Err),
			)
		} else {
			l.logDebugEvent("OnStart hook executed",
				slog.String("callee", e.FunctionName),
				slog.String("caller", e.CallerName),
				slog.String("runtime", e.Runtime.String()),
			)
		}
	case *fxevent.OnStopExecuting:
		l.logDebugEvent("OnStop hook executing",
			slog.String("callee", e.FunctionName),
			slog.String("caller", e.CallerName),
		)
	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			l.logError("OnStop hook failed",
				slog.String("callee", e.FunctionName),
				slog.String("caller", e.CallerName),
				slogErr(e.Err),
			)
		} else {
			l.logDebugEvent("OnStop hook executed",
				slog.String("callee", e.FunctionName),
				slog.String("caller", e.CallerName),
				slog.String("runtime", e.Runtime.String()),
			)
		}
	case *fxevent.Supplied:
		if e.Err != nil {
			l.logError("error encountered while applying options",
				slog.String("type", e.TypeName),
				slogStrings("moduletrace", e.ModuleTrace),
				slogStackTrace("stacktrace", e.StackTrace),
				slogMaybeModuleField(e.ModuleName),
				slogErr(e.Err))
		} else {
			l.logEvent("supplied",
				slog.String("type", e.TypeName),
				slogStackTrace("stacktrace", e.StackTrace),
				slogStrings("moduletrace", e.ModuleTrace),
				slogMaybeModuleField(e.ModuleName),
			)
		}
	case *fxevent.Provided:
		if e.Err != nil {
			for _, rtype := range e.OutputTypeNames {
				l.logError("provided",
					slog.String("type", rtype),
					slog.String("constructor", e.ConstructorName),
				)
			}
			l.logError(strings.Join(e.StackTrace, "\n"))
		} else {
			for _, rtype := range e.OutputTypeNames {
				l.logInfoEvent("provided",
					slog.String("type", rtype),
					slog.String("constructor", e.ConstructorName),
				)
			}
		}
	case *fxevent.Replaced:
		for _, rtype := range e.OutputTypeNames {
			l.logEvent("replaced",
				slogStackTrace("stacktrace", e.StackTrace),
				slogStrings("moduletrace", e.ModuleTrace),
				slogMaybeModuleField(e.ModuleName),
				slog.String("type", rtype),
			)
		}
		if e.Err != nil {
			l.logError("error encountered while replacing",
				slogStackTrace("stacktrace", e.StackTrace),
				slogStrings("moduletrace", e.ModuleTrace),
				slogMaybeModuleField(e.ModuleName),
				slogErr(e.Err))
		}
	case *fxevent.Decorated:
		if e.Err != nil {
			for _, rtype := range e.OutputTypeNames {
				l.logError("prodecoratedvided",
					slog.String("type", rtype),
					slog.String("decorator", e.DecoratorName),
				)
			}
			l.logError(strings.Join(e.StackTrace, "\n"))
		} else {
			for _, rtype := range e.OutputTypeNames {
				l.logInfoEvent("decorated",
					slog.String("type", rtype),
					slog.String("decorator", e.DecoratorName),
				)
			}
		}
	case *fxevent.BeforeRun:
		l.logDebugEvent("before run",
			slog.String("name", e.Name),
			slog.String("kind", e.Kind),
			slogMaybeModuleField(e.ModuleName),
		)
	case *fxevent.Run:
		if e.Err != nil {
			l.logError("error returned",
				slog.String("name", e.Name),
				slog.String("kind", e.Kind),
				slogMaybeModuleField(e.ModuleName),
				slogErr(e.Err),
			)
		} else {
			l.logDebugEvent("run",
				slog.String("name", e.Name),
				slog.String("kind", e.Kind),
				slog.String("runtime", e.Runtime.String()),
				slogMaybeModuleField(e.ModuleName),
			)
		}
	case *fxevent.Invoking:
		// Do not log stack as it will make logs hard to read.
		l.logDebugEvent("invoking",
			slog.String("function", e.FunctionName),
			slogMaybeModuleField(e.ModuleName),
		)
	case *fxevent.Invoked:
		if e.Err != nil {
			l.logError(fmt.Sprintf("invoke failed: %s\n%s", e.Err, e.Trace)) // slog.String("function", e.FunctionName),
			// slogMaybeModuleField(e.ModuleName),

		}
	case *fxevent.Stopping:
		l.logEvent("received signal",
			slog.String("signal", strings.ToUpper(e.Signal.String())))
	case *fxevent.Stopped:
		if e.Err != nil {
			l.logError("stop failed", slogErr(e.Err))
		}
	case *fxevent.RollingBack:
		l.logError("start failed, rolling back", slogErr(e.StartErr))
	case *fxevent.RolledBack:
		if e.Err != nil {
			l.logError("rollback failed", slogErr(e.Err))
		}
	case *fxevent.Started:
		if e.Err != nil {
			l.logError("start failed", slogErr(e.Err))
		} else {
			l.logDebugEvent("started")
		}
	case *fxevent.LoggerInitialized:
		if e.Err != nil {
			l.logError("custom logger initialization failed", slogErr(e.Err))
		} else {
			l.logEvent("initialized custom fxevent.Logger", slog.String("function", e.ConstructorName))
		}
	}
}

type slogFieldSkip struct{}

func slogMaybeModuleField(name string) slog.Attr {
	if len(name) == 0 {
		return slog.Any("module", slogFieldSkip{})
	}
	return slog.String("module", name)
}

func slogMaybeBool(name string, b bool) slog.Attr {
	if !b {
		return slog.Any(name, slogFieldSkip{})
	}
	return slog.Bool(name, true)
}

func slogErr(err error) slog.Attr {
	return slog.String("error", err.Error())
}

func slogStrings(key string, str []string) slog.Attr {
	attrs := make([]any, len(str))
	for i, val := range str {
		attrs[i] = slog.String(strconv.Itoa(i), val)
	}
	return slog.Group(key, attrs...)
}

func slogStackTrace(key string, str []string) slog.Attr {
	return slog.String(key, strings.Join(str, "\n"))
}
