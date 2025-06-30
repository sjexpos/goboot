package log

import (
	"context"
	"fmt"
	"github.com/sjexpos/goboot/concurrent"
	"log/slog"
)

const GO_ROUTINE_ID_FIELD_NAME = "GID"
const GO_ROUTINE_NAME_FIELD_NAME = "GNAME"
const APP_ID_FIELD_NAME = "APPID"

type localHandler slog.Handler

func NewSlogEnhancedHandler(appId string, handler slog.Handler, goIdFormat string, goNameFormat string) slog.Handler {
	h := &slogEnhancedHandler{}
	h.appId = appId
	h.localHandler = handler
	h.goIdFormat = goIdFormat
	h.goNameFormat = goNameFormat
	return h
}

type slogEnhancedHandler struct {
	localHandler
	appId        string
	goIdFormat   string
	goNameFormat string
}

func (s *slogEnhancedHandler) Handle(ctx context.Context, record slog.Record) error {
	gid := fmt.Sprintf(s.goIdFormat, concurrent.GoroutineID())
	record.Add(slog.String(GO_ROUTINE_ID_FIELD_NAME, gid))
	if s.appId != "" {
		record.Add(slog.String(APP_ID_FIELD_NAME, s.appId))
	}
	p := MDC.allValues()
	if p != nil {
		values := *p
		for k, v := range values {
			if k == GO_ROUTINE_NAME_FIELD_NAME {
				gname := fmt.Sprintf(s.goNameFormat, v)
				record.Add(slog.String(k, gname))
			} else {
				record.Add(slog.String(k, v))
			}
		}
		_, found := values[GO_ROUTINE_NAME_FIELD_NAME]
		if !found {
			gname := fmt.Sprintf(s.goNameFormat, concurrent.GoroutineID())
			record.Add(slog.String(GO_ROUTINE_NAME_FIELD_NAME, gname))
		}
	}
	return s.localHandler.Handle(ctx, record)
}

func (s *slogEnhancedHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handler := s.localHandler.WithAttrs(attrs)
	h := &slogEnhancedHandler{}
	h.appId = s.appId
	h.goIdFormat = s.goIdFormat
	h.goNameFormat = s.goNameFormat
	h.localHandler = handler
	return h
}

func (s *slogEnhancedHandler) WithGroup(name string) slog.Handler {
	handler := s.localHandler.WithGroup(name)
	h := &slogEnhancedHandler{}
	h.appId = s.appId
	h.goIdFormat = s.goIdFormat
	h.goNameFormat = s.goNameFormat
	h.localHandler = handler
	return h
}
