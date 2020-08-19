package log

import (
	"github.com/sirupsen/logrus"
	"minGateway/util/log/logrushook"
	"time"
)

var L *logrus.Logger

func Init(level string) {
	L = logrus.New()
	L.SetLevel(getLogLevel(level))
}

func WirteLog(logPath string) {
	//分割日志，每小时一个文件，最长保存15天
	lfsHook := logrushook.NewLfsHook(logPath, time.Hour*24*15, time.Hour)
	L.AddHook(lfsHook)
}

//设置最低loglevel: debug,info,warn,error,fatal,panic
func getLogLevel(levelName string) logrus.Level {
	switch levelName {
	case "debug":
		return logrus.DebugLevel
	case "DebugLevel":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "InfoLevel":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "WarnLevel":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "ErrorLevel":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	case "FatalLevel":
		return logrus.FatalLevel
	case "panic":
		return logrus.PanicLevel
	case "PanicLevel":
		return logrus.PanicLevel
	default:
		return logrus.WarnLevel
	}
}

func Tracef(format string, args ...interface{}) {
	L.Tracef(format, args...)
}

func Debugf(format string, args ...interface{}) {
	L.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	L.Infof(format, args...)
}

func Printf(format string, args ...interface{}) {
	L.Printf(format, args...)
}

func Warnf(format string, args ...interface{}) {
	L.Warnf(format, args...)
}

func Warningf(format string, args ...interface{}) {
	L.Warningf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	L.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	L.Fatalf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	L.Panicf(format, args...)
}

func Trace(args ...interface{}) {
	L.Trace(args...)
}

func Debug(args ...interface{}) {
	L.Debug(args...)
}

func Info(args ...interface{}) {
	L.Info(args...)
}

func Print(args ...interface{}) {
	L.Print(args...)
}

func Warn(args ...interface{}) {
	L.Warn(args...)
}

func Warning(args ...interface{}) {
	L.Warning(args...)
}

func Error(args ...interface{}) {
	L.Error(args...)
}

func Fatal(args ...interface{}) {
	L.Fatal(args...)
}

func Panic(args ...interface{}) {
	L.Panic(args...)
}
