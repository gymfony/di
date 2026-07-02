package digen

import "fmt"

type IParserLogger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

type ParserLogger struct {
	logger IParserLogger
}

func NewParserLogger(logger IParserLogger) *ParserLogger {
	if logger == nil {
		return &ParserLogger{logger: nil}
	}
	return &ParserLogger{logger: logger}
}

func (l *ParserLogger) logDebug(msg string, args ...any) {
	if l == nil || l.logger == nil {
		return
	}
	if len(args) > 0 {
		l.logger.Debug(fmt.Sprintf(msg, args...))
	} else {
		l.logger.Debug(msg)
	}
}

func (l *ParserLogger) logInfo(msg string, args ...any) {
	if l == nil || l.logger == nil {
		return
	}
	if len(args) > 0 {
		l.logger.Info(fmt.Sprintf(msg, args...))
	} else {
		l.logger.Info(msg)
	}
}

func (l *ParserLogger) logWarn(msg string, args ...any) {
	if l == nil || l.logger == nil {
		return
	}
	if len(args) > 0 {
		l.logger.Warn(fmt.Sprintf(msg, args...))
	} else {
		l.logger.Warn(msg)
	}
}

func (l *ParserLogger) logError(msg string, args ...any) {
	if l == nil || l.logger == nil {
		return
	}
	if len(args) > 0 {
		l.logger.Error(fmt.Sprintf(msg, args...))
	} else {
		l.logger.Error(msg)
	}
}
