package myapi

import (
	"io"
	"sync/atomic"

	elog "github.com/labstack/gommon/log"
	"github.com/recallsong/go-utils/encoding/jsonx"
	logrus "github.com/sirupsen/logrus"
)

var stdLogger = logrus.StandardLogger()

type LogrusToEchoLogger struct{}

func (l LogrusToEchoLogger) Output() io.Writer {
	return stdLogger.Out
}
func (l LogrusToEchoLogger) SetOutput(w io.Writer) {
	logrus.SetOutput(w)
}
func (l LogrusToEchoLogger) Prefix() string {
	return ""
}
func (l LogrusToEchoLogger) SetPrefix(p string) {
	return
}
func (l LogrusToEchoLogger) Level() elog.Lvl {
	switch logrus.GetLevel() {
	case logrus.DebugLevel:
		return elog.DEBUG
	case logrus.InfoLevel:
		return elog.INFO
	case logrus.WarnLevel:
		return elog.WARN
	case logrus.ErrorLevel:
		return elog.ERROR
	}
	return elog.OFF
}
func (l LogrusToEchoLogger) SetLevel(v elog.Lvl) {
	switch v {
	case elog.DEBUG:
		logrus.SetLevel(logrus.DebugLevel)
	case elog.INFO:
		logrus.SetLevel(logrus.InfoLevel)
	case elog.WARN:
		logrus.SetLevel(logrus.WarnLevel)
	case elog.ERROR:
		logrus.SetLevel(logrus.ErrorLevel)
	}
}
func (l LogrusToEchoLogger) Print(i ...interface{}) {
	logrus.Print(i...)
}
func (l LogrusToEchoLogger) Printf(format string, args ...interface{}) {
	logrus.Printf(format, args...)
}
func (l LogrusToEchoLogger) Printj(j elog.JSON) {
	logrus.Print(jsonx.Marshal(j))
}
func (l LogrusToEchoLogger) Debug(i ...interface{}) {
	logrus.Debug(i...)
}
func (l LogrusToEchoLogger) Debugf(format string, args ...interface{}) {
	logrus.Debugf(format, args)
}
func (l LogrusToEchoLogger) Debugj(j elog.JSON) {
	if logrus.Level(atomic.LoadUint32((*uint32)(&stdLogger.Level))) >= logrus.DebugLevel {
		logrus.Debug(jsonx.Marshal(j))
	}
}
func (l LogrusToEchoLogger) Info(i ...interface{}) {
	logrus.Info(i...)
}
func (l LogrusToEchoLogger) Infof(format string, args ...interface{}) {
	logrus.Infof(format, args)
}
func (l LogrusToEchoLogger) Infoj(j elog.JSON) {
	if logrus.Level(atomic.LoadUint32((*uint32)(&stdLogger.Level))) >= logrus.InfoLevel {
		logrus.Info(jsonx.Marshal(j))
	}
}
func (l LogrusToEchoLogger) Warn(i ...interface{}) {
	logrus.Warn(i...)
}
func (l LogrusToEchoLogger) Warnf(format string, args ...interface{}) {
	logrus.Warnf(format, args)
}
func (l LogrusToEchoLogger) Warnj(j elog.JSON) {
	if logrus.Level(atomic.LoadUint32((*uint32)(&stdLogger.Level))) >= logrus.WarnLevel {
		logrus.Warn(jsonx.Marshal(j))
	}
}
func (l LogrusToEchoLogger) Error(i ...interface{}) {
	logrus.Error(i...)
}
func (l LogrusToEchoLogger) Errorf(format string, args ...interface{}) {
	logrus.Errorf(format, args)
}
func (l LogrusToEchoLogger) Errorj(j elog.JSON) {
	if logrus.Level(atomic.LoadUint32((*uint32)(&stdLogger.Level))) >= logrus.ErrorLevel {
		logrus.Error(jsonx.Marshal(j))
	}
}
func (l LogrusToEchoLogger) Fatal(i ...interface{}) {
	logrus.Fatal(i...)
}
func (l LogrusToEchoLogger) Fatalj(j elog.JSON) {
	if logrus.Level(atomic.LoadUint32((*uint32)(&stdLogger.Level))) >= logrus.FatalLevel {
		logrus.Fatal(jsonx.Marshal(j))
	}
}
func (l LogrusToEchoLogger) Fatalf(format string, args ...interface{}) {
	logrus.Fatalf(format, args)
}
func (l LogrusToEchoLogger) Panic(i ...interface{}) {
	logrus.Panic(i...)
}
func (l LogrusToEchoLogger) Panicj(j elog.JSON) {
	if logrus.Level(atomic.LoadUint32((*uint32)(&stdLogger.Level))) >= logrus.PanicLevel {
		logrus.Panic(jsonx.Marshal(j))
	}
}
func (l LogrusToEchoLogger) Panicf(format string, args ...interface{}) {
	logrus.Panicf(format, args)
}
