package logger

import (
	"github.com/xacnio/backupper/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

type LogWriteType uint

type LogConfig struct {
	Output string
	Name   string
	Level  zapcore.Level
}

var (
	Main  *zap.SugaredLogger
	SSH   *zap.SugaredLogger
	SFTP  *zap.SugaredLogger
	FTP   *zap.SugaredLogger
	TgBot *zap.SugaredLogger
)

var Logs = []LogConfig{
	{Output: "main.log", Name: "Main"},
	{Output: "ssh.log", Name: "SSH"},
	{Output: "sftp.log", Name: "SFTP"},
	{Output: "ftp.log", Name: "FTP"},
	{Output: "tgbot.log", Name: "TgBot"},
}

func loggerConfigBuilder(lc LogConfig) zap.Config {
	zapConfig := zap.NewProductionConfig()
	zapConfig.OutputPaths = []string{
		"logs/" + lc.Output,
	}
	if lc.Name == "Main" {
		zapConfig.OutputPaths = append(zapConfig.OutputPaths, "stdout")
	}
	zapConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("Jan 02 15:04:05.000000000")
	logLevel := "info"
	if config.Get().LogLevel != nil {
		logLevel = *config.Get().LogLevel
	}
	switch logLevel {
	case "debug":
		zapConfig.Level.SetLevel(zap.DebugLevel)
	case "info":
		zapConfig.Level.SetLevel(zap.InfoLevel)
	case "warn":
		zapConfig.Level.SetLevel(zap.WarnLevel)
	case "error":
		zapConfig.Level.SetLevel(zap.ErrorLevel)
	case "fatal":
		zapConfig.Level.SetLevel(zap.FatalLevel)
	case "dpanic":
		zapConfig.Level.SetLevel(zap.DPanicLevel)
	case "panic":
		zapConfig.Level.SetLevel(zap.PanicLevel)
	default:
		zapConfig.Level.SetLevel(zap.InfoLevel)
	}
	return zapConfig
}

func Init() {
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		_ = os.Mkdir("logs", os.ModePerm)
	}

	for _, log := range Logs {
		_logger, err := loggerConfigBuilder(log).Build()
		if err != nil {
			panic(err)
		}
		_sugar := _logger.Sugar()
		switch log.Name {
		case "Main":
			Main = _sugar
		case "SSH":
			SSH = _sugar
		case "SFTP":
			SFTP = _sugar
		case "FTP":
			FTP = _sugar
		case "TgBot":
			TgBot = _sugar
		}
	}
}
