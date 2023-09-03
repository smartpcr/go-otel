package ot

import (
	"context"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

type otLogrusHook struct {
	Writer    *lumberjack.Logger
	LogLevels []logrus.Level
}

var _ logrus.Hook = otLogrusHook{}

func RegisterLogger(ctx context.Context) *logrus.Entry {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	fileLogger := &lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   // days
		Compress:   true, // disabled by default
	}
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.TraceLevel)
	logger.AddHook(&otLogrusHook{
		Writer:    fileLogger,
		LogLevels: logrus.AllLevels,
	})
	return logrus.WithContext(ctx)
}

func (o otLogrusHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (o otLogrusHook) Fire(entry *logrus.Entry) error {
	ctx := entry.Context
	if ctx == nil {
		return nil
	}
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return nil
	}

	spanCtx := span.SpanContext()
	if spanCtx.HasTraceID() {
		entry.Data["trace_id"] = spanCtx.TraceID().String()
	}
	if spanCtx.HasSpanID() {
		entry.Data["span_id"] = spanCtx.SpanID().String()
	}

	attrs := make([]attribute.KeyValue, 0)
	logSeverityKey := attribute.Key("log.severity")
	logMessageKey := attribute.Key("log.message")
	attrs = append(attrs, logSeverityKey.String(entry.Level.String()))
	attrs = append(attrs, logMessageKey.String(entry.Message))
	span.AddEvent("log", trace.WithAttributes(attrs...))

	if entry.Level < logrus.ErrorLevel {
		span.SetStatus(codes.Error, entry.Message)
	}

	return nil
}
