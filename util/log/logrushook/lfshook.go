package logrushook

import (
	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/pkg/errors"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
	"time"
)

// WithMaxAge和WithRotationCount二者只能设置一个，
// WithMaxAge设置文件清理前的最长保存时间，
// WithRotationCount设置文件清理前最多保存的个数。
// rotatelogs.WithMaxAge(time.Hour*24),
func NewLfsHook(logPath string, maxAge time.Duration, rotationTime time.Duration) log.Hook {
	writer, err := rotatelogs.New(
		logPath+".%Y-%m-%d-%H-%M",
		rotatelogs.WithLinkName(logPath),          // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(maxAge),             // 文件最大保存时间
		rotatelogs.WithRotationTime(rotationTime), // 日志切割时间间隔
	)
	if err != nil {
		log.Errorf("config local file system logger error. %+v", errors.WithStack(err))
	}
	lfsHook := lfshook.NewHook(lfshook.WriterMap{
		log.DebugLevel: writer,
		log.InfoLevel:  writer,
		log.WarnLevel:  writer,
		log.ErrorLevel: writer,
		log.FatalLevel: writer,
		log.PanicLevel: writer,
	}, &log.TextFormatter{DisableColors: true})

	return lfsHook
}
