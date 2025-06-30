package sentry

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/yoshino-s/go-framework/application"
	"go.uber.org/zap/zapcore"
)

var _ zapcore.Core = (*sentryLoggerCore)(nil)

type sentryLoggerCore struct {
	zapcore.LevelEnabler                        // Determines which log levels are enabled
	fields               map[string]interface{} // Additional fields to include with each log entry
	context              context.Context        // Context for Sentry operations, may contain a Sentry span
}

var _ application.LoggerCoreContrib = (*Sentry)(nil)

func (c *Sentry) LoggerCores() []zapcore.Core {
	return []zapcore.Core{
		&sentryLoggerCore{
			LevelEnabler: zapcore.ErrorLevel,
			fields:       make(map[string]interface{}),
			context:      context.Background(),
		},
	}
}

func (s *sentryLoggerCore) With(fields []zapcore.Field) zapcore.Core {
	return s.addFields(fields)
}

func (s *sentryLoggerCore) addFields(fields []zapcore.Field) *sentryLoggerCore {
	// Start with the current context or a background context if none exists
	currentContext := s.context
	if currentContext == nil {
		currentContext = context.Background()
	}

	// Copy existing fields
	m := make(map[string]interface{}, len(s.fields))
	for k, v := range s.fields {
		m[k] = v
	}

	// Add fields to an in-memory encoder
	enc := zapcore.NewMapObjectEncoder()

	for _, f := range fields {
		// Extract context if present
		if v, ok := f.Interface.(context.Context); ok && v != nil {
			currentContext = v
			continue
		}

		// Add non-skip fields to the encoder
		if f.Type != zapcore.SkipType {
			f.AddTo(enc)
		}
	}

	// Merge the encoded fields into our map
	for k, v := range enc.Fields {
		m[k] = v
	}

	// Create a new core with the updated fields and context
	return &sentryLoggerCore{
		LevelEnabler: s.LevelEnabler,
		fields:       m,
		context:      currentContext,
	}
}

// Check determines whether the supplied Entry should be logged.
// It implements zapcore.Core interface.
func (s *sentryLoggerCore) Check(entry zapcore.Entry, checkEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if s.Enabled(entry.Level) {
		return checkEntry.AddCore(entry, s)
	}

	return checkEntry
}

// flushSentry flushes any buffered Sentry events with the given timeout
func flushSentry() {
	sentry.Flush(2 * time.Second)
}

// Write takes a log entry and sends it to Sentry.
// It implements zapcore.Core interface.
func (s *sentryLoggerCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	// Flush Sentry events when the function returns
	defer flushSentry()

	// Create a clone with the additional fields
	clone := s.addFields(fields)

	// Extract span from context if present
	span := sentry.SpanFromContext(clone.context)

	// Create a local hub to avoid modifying the global hub
	localHub := sentry.CurrentHub().Clone()

	// Get the Sentry client
	client := localHub.Client()
	if client == nil {
		// No client configured, nothing to do
		return nil
	}

	// Configure the scope with caller information and span
	localHub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag("file", entry.Caller.File)
		scope.SetTag("line", strconv.Itoa(entry.Caller.Line))
		scope.SetSpan(span)
	})

	errors := []error{}
	for _, field := range fields {
		if field.Type == zapcore.ErrorType {
			// If the field is an error, add it to the extra fields
			if err, ok := field.Interface.(error); ok {
				if err != nil {
					errors = append(errors, err)
				}
			}
		}
	}

	// Create the Sentry event
	event := &sentry.Event{
		Extra:       clone.fields,
		Fingerprint: []string{entry.Message},
		Level:       sentrySeverity(entry.Level),
		Message:     entry.Message,
		Platform:    "go",
		Timestamp:   entry.Time,
		Logger:      entry.LoggerName,
	}

	// Add exception with stack trace for error-level logs if enabled
	for _, err := range errors {
		event.SetException(err, client.Options().MaxErrorDepth)
	}

	fmt.Println("Sending Sentry event:", event.Message, "at", event.Timestamp)
	// Send the event to Sentry
	client.CaptureEvent(event, nil, localHub.Scope())

	return nil
}

// Sync flushes any buffered log entries.
// It implements zapcore.Core interface.
func (*sentryLoggerCore) Sync() error {
	flushSentry()
	return nil
}

// sentrySeverity converts a Zap log level to the corresponding Sentry level.
// This ensures that log levels are properly mapped between the two systems.
func sentrySeverity(lvl zapcore.Level) sentry.Level {
	switch lvl {
	case zapcore.DebugLevel:
		return sentry.LevelDebug
	case zapcore.InfoLevel:
		return sentry.LevelInfo
	case zapcore.WarnLevel:
		return sentry.LevelWarning
	case zapcore.ErrorLevel:
		return sentry.LevelError
	case zapcore.DPanicLevel:
		return sentry.LevelFatal
	case zapcore.PanicLevel:
		return sentry.LevelFatal
	case zapcore.FatalLevel:
		return sentry.LevelFatal
	default:
		// Unrecognized levels are treated as fatal to ensure they're noticed
		return sentry.LevelFatal
	}
}
