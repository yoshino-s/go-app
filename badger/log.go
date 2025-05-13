package badger

import (
	"github.com/dgraph-io/badger/v4"
	"go.uber.org/zap"
)

var _ badger.Logger = &logger{}

type logger struct {
	*zap.SugaredLogger
}

// Warningf implements badger.Logger.
func (l *logger) Warningf(msg string, args ...interface{}) {
	l.Warnf(msg, args...)
}
