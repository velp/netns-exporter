package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a global container for the initialized logger.
var Logger *zap.SugaredLogger

// InitLoggerOpts represents options for logger.
type InitLoggerOpts struct {
	UseStdout bool
	Debug     bool
	File      string
}

// InitLogger initializes a global logger for the application.
func InitLogger(opts InitLoggerOpts) error {
	// Configure loglevel.
	var loglevel zap.AtomicLevel
	if opts.Debug {
		loglevel = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else {
		loglevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	outputPaths := []string{}
	if opts.File != "" {
		outputPaths = append(outputPaths, opts.File)
	}
	if opts.UseStdout {
		outputPaths = append(outputPaths, "stdout")
	}

	// Configure the zap logger.
	cfg := zap.Config{
		Level:    loglevel,
		Encoding: "console",
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:     "ts",
			LevelKey:    "level",
			MessageKey:  "msg",
			EncodeLevel: zapcore.CapitalColorLevelEncoder,
			EncodeTime:  zapcore.ISO8601TimeEncoder,
		},
		OutputPaths:      outputPaths,
		ErrorOutputPaths: outputPaths,
	}
	logger, err := cfg.Build()
	if err != nil {
		return err
	}

	zap.ReplaceGlobals(logger)

	Logger = zap.S()
	return nil
}
