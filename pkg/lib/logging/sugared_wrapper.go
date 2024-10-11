package logging

import (
	"go.uber.org/zap"

	"github.com/anyproto/anytype-heart/util/anyerror"
)

type Sugared struct {
	*zap.SugaredLogger
}

func (s *Sugared) With(args ...interface{}) *Sugared {
	cleanupArgs(args)
	return &Sugared{s.SugaredLogger.With(args...)}
}

func (s *Sugared) Warn(args ...interface{}) {
	cleanupArgs(args)
	s.SugaredLogger.Warn(args...)
}

func (s *Sugared) Warnf(template string, args ...interface{}) {
	cleanupArgs(args)
	s.SugaredLogger.Warnf(template, args...)
}

func (s *Sugared) Warnw(msg string, keysAndValues ...interface{}) {
	cleanupArgs(keysAndValues)
	s.SugaredLogger.Warnw(msg, keysAndValues...)
}

func (s *Sugared) Error(args ...interface{}) {
	cleanupArgs(args)
	s.SugaredLogger.Error(args...)
}

func (s *Sugared) Errorf(template string, args ...interface{}) {
	cleanupArgs(args)
	s.SugaredLogger.Errorf(template, args...)
}

func (s *Sugared) Errorw(msg string, keysAndValues ...interface{}) {
	cleanupArgs(keysAndValues)
	s.SugaredLogger.Errorw(msg, keysAndValues...)
}

func (s *Sugared) Info(args ...interface{}) {
	cleanupArgs(args)
	s.SugaredLogger.Info(args...)
}

func (s *Sugared) Infof(template string, args ...interface{}) {
	cleanupArgs(args)
	s.SugaredLogger.Infof(template, args...)
}

func (s *Sugared) Infow(msg string, keysAndValues ...interface{}) {
	cleanupArgs(keysAndValues)
	s.SugaredLogger.Infow(msg, keysAndValues...)
}

func cleanupArgs(args []interface{}) {
	for i, arg := range args {
		if err, ok := arg.(error); ok {
			err = anyerror.CleanupError(err)
			args[i] = err
		}
	}
}
