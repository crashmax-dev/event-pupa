package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

var (
	isSinkRegistered bool
)

func Initialize(level zapcore.Level, path string, postfix string) (*zap.SugaredLogger, *zap.AtomicLevel) {

	if path == "" {
		path = "logs"
	}

	var (
		levelSelected zapcore.Level
	)

	levelSelected = NormalizeLevel(level)

	atom := zap.NewAtomicLevelAt(levelSelected)

	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Println(err)
	}

	filename := getOSFilePath(filepath.Join(path,
		fmt.Sprintf("log_%s%s.log", postfix,
			time.Now().Format("02012006"))))

	config := zap.Config{
		Level:    atom,
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:     "time",
			MessageKey:  "message",
			LevelKey:    "level",
			NameKey:     "namekey",
			EncodeLevel: zapcore.LowercaseLevelEncoder,
			EncodeTime:  zapcore.ISO8601TimeEncoder},
		OutputPaths:      []string{filename},
		ErrorOutputPaths: []string{filename},
	}
	newWinFileSink := func(u *url.URL) (zap.Sink, error) {
		// Remove leading slash left by url.Parse()
		return os.OpenFile(u.Path[1:], os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	}

	//rawJSON := []byte(`{
	//  "level": "debug",
	//  "encoding": "json",
	//  "encoderConfig": {
	//    "messageKey": "message",
	//    "levelKey": "level",
	//    "levelEncoder": "lowercase"
	//  }
	//}`)

	if isSinkRegistered == false {
		err = zap.RegisterSink("winfile", newWinFileSink)
		if err != nil {
			panic(err)
		}
		isSinkRegistered = true
	}

	logger := zap.Must(config.Build())

	defer syncLogger(logger)

	logger.Info("logger construction succeeded")

	return logger.Sugar(), &atom
}

func syncLogger(logger *zap.Logger) {
	err := logger.Sync()
	if err != nil {
		panic(err)
	}
}

func NormalizeLevel(level zapcore.Level) zapcore.Level {
	if level == zap.DebugLevel {
		return zapcore.DebugLevel
	} else {
		return zapcore.ErrorLevel
	}
}
