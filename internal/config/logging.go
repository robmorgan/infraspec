package config

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type logging struct {
	Logger          *zap.SugaredLogger
	AtomicLogLevel  zap.AtomicLevel
	DefaultLogLevel zapcore.Level
}

var (
	logger *zap.Logger
	log    *zap.SugaredLogger

	// Logging is the public interface to logging
	Logging = &logging{
		AtomicLogLevel:  zap.NewAtomicLevel(),
		DefaultLogLevel: zap.InfoLevel,
	}
)

// init creates a logger with custom encoding config
func init() {
	Logging.AtomicLogLevel = zap.NewAtomicLevel()
	// zap needs to start at zapcore.DebugLevel so that it can then be decreased to a lesser level
	Logging.AtomicLogLevel.SetLevel(zapcore.DebugLevel)
	encoderCfg := zap.NewProductionEncoderConfig()

	encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderCfg.EncodeDuration = nil
	encoderCfg.EncodeTime = nil
	//encoderCfg.TimeKey = ""
	encoderCfg.EncodeCaller = nil

	logger = zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		Logging.AtomicLogLevel,
	))

	defer logger.Sync() // flushes buffer, if any
	log = logger.Sugar()
	Logging.Logger = log

	//cfg.DisableStacktrace = true
	//cfg.EncoderConfig.EncodeCaller = nil
	// if os.Getenv("INFRASPEC_DEBUG") != "" {
	// 	cfg.DisableStacktrace = false
	// 	cfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	// }
}

func (logging) setLogLevel(lvl zapcore.Level) {
	if Logging.AtomicLogLevel.Level() != lvl {
		log.Infof("setting LogLevel to %s", lvl)
		Logging.AtomicLogLevel.SetLevel(lvl)
	}
}
