package gorm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	glogger "gorm.io/gorm/logger"
)

func NewSLog(logLevel glogger.LogLevel, slowThreshold time.Duration) glogger.Interface {
	return &logger{
		logger:                    slog.With().WithGroup("GORM"),
		SlowThreshold:             slowThreshold,
		IgnoreRecordNotFoundError: true,
		LogLevel:                  logLevel,
	}
}

func NewSLog2(level string, slowThreshold time.Duration) glogger.Interface {
	var logLevel glogger.LogLevel
	if strings.EqualFold(level, "Error") {
		logLevel = glogger.Error
	} else if strings.EqualFold(level, "Warn") {
		logLevel = glogger.Warn
	} else if strings.EqualFold(level, "Info") {
		logLevel = glogger.Info
	} else {
		logLevel = glogger.Silent
	}
	return &logger{
		logger:                    slog.With().WithGroup("GORM"),
		SlowThreshold:             slowThreshold,
		IgnoreRecordNotFoundError: true,
		LogLevel:                  logLevel,
	}
}

type logger struct {
	logger                    *slog.Logger
	SlowThreshold             time.Duration
	Colorful                  bool
	IgnoreRecordNotFoundError bool
	ParameterizedQueries      bool
	LogLevel                  glogger.LogLevel
}

func (l *logger) LogMode(level glogger.LogLevel) glogger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

func (l *logger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= glogger.Info {
		l.logger.Info(fmt.Sprintf(msg, data...))
	}
}

func (l *logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= glogger.Warn {
		l.logger.Warn(fmt.Sprintf(msg, data...))
	}
}

func (l *logger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= glogger.Error {
		l.logger.Error(fmt.Sprintf(msg, data...))
	}
}

func (l *logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	str := l.printLine(begin, fc, err)
	l.logger.Debug(str)
}

func (l *logger) printLine(begin time.Time, fc func() (string, int64), err error) string {
	var (
		traceStr     string = glogger.Red + "[%.3fms] " + glogger.Green + "[rows:%v]" + glogger.Reset + " %s"
		traceWarnStr string = glogger.Yellow + "%s" + glogger.Reset + glogger.RedBold + "[%.3fms] " + glogger.Yellow + "[rows:%v]" + glogger.Magenta + " %s" + glogger.Reset
		traceErrStr  string = glogger.RedBold + "%s " + glogger.Reset + glogger.Yellow + "[%.3fms] " + glogger.BlueBold + "[rows:%v]" + glogger.Reset + " %s"
	)
	elapsed := time.Since(begin)
	sql, rows := fc()
	switch {
	case err != nil && (!errors.Is(err, glogger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		if rows == -1 {
			return fmt.Sprintf(traceErrStr, err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			return fmt.Sprintf(traceErrStr, err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= glogger.Warn:
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			return fmt.Sprintf(traceWarnStr, slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			return fmt.Sprintf(traceWarnStr, slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case l.LogLevel == glogger.Info:
		if rows == -1 {
			return fmt.Sprintf(traceStr, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			return fmt.Sprintf(traceStr, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
	return ""
}
