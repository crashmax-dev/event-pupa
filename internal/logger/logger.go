package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

var (
	isSinkRegistered bool
)

// Initialize инициализирует логгер с уровнем логгирования level, в папке path по относительному пути, с добавлением
// postfix к имени файла
func Initialize(level zapcore.Level, path string, postfix string) (*zap.SugaredLogger, *zap.AtomicLevel, error) {

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
		fmt.Println(err)
		return nil, nil, err
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

	if isSinkRegistered == false {
		err = zap.RegisterSink("winfile", newWinFileSink)
		if err != nil {
			return nil, nil, err
		}
		isSinkRegistered = true
	}

	logger := zap.Must(config.Build())

	logger.Info("logger construction succeeded")

	errSync := logger.Sync()
	if errSync != nil {
		fmt.Println("logger sync failed: ", errSync)
		return nil, nil, errSync
	}

	return logger.Sugar(), &atom, nil
}

// NormalizeLevel выравнивает уровень для Dev и Prod, возвращая DebugLevel или ErrorLevel соответственно.
func NormalizeLevel(level zapcore.Level) zapcore.Level {
	if level == zap.DebugLevel {
		return zapcore.DebugLevel
	} else {
		return zapcore.ErrorLevel
	}
}
