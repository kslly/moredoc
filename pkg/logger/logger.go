package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	timeFormat = "060102 15:04:05.0000"
)

var (
	otputPaths = []string{"stderr"}
	Vicglog, _ = createGlobalLogger() // VICG日志
)

// Level 声明日志级别
type Level zapcore.Level

const (
	DebugLevel Level = iota - 1
	InfoLevel
	WarnLevel
	ErrorLevel
	// dPanicLevel // 占位, 不给外部使用
	// panicLevel  // 占位, 不给外部使用
	FatalLevel = 5

	DEBUG = "DEBUG"
)

// Logger 定义logger接口
type Logger interface {
	SetLevel(Level)
	GetLevel() string
	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
	Fatalf(string, ...interface{})
}

// AppLogger 实现一个logger
type AppLogger struct {
	cfg   *zap.Config
	slg   *zap.SugaredLogger
	lg    *zap.Logger
	level zap.AtomicLevel
}

// SetLevel: 维护日志级别的方法
// 使用一个配置接口维护这个方法
func (applg *AppLogger) SetLevel(l Level) {
	applg.level.SetLevel(zapcore.Level(l))
}

// GetLevel 获取logger的等级
func (applg *AppLogger) GetLevel() string {
	return applg.level.Level().CapitalString()
}

// Debug logs a message at DebugLevel with SugaredLogger
func (applg *AppLogger) Debugf(temp string, args ...interface{}) {
	applg.slg.Debugf(temp, args...)
}

func (applg *AppLogger) IsDebugLevel() bool {
	return applg.level.Level() == zapcore.Level(DebugLevel)
}

// Infof logs a message at InfoLevel with SugaredLogger
func (applg *AppLogger) Infof(temp string, args ...interface{}) {
	applg.slg.Infof(temp, args...)
}

func (applg *AppLogger) Warnf(temp string, args ...interface{}) {
	applg.slg.Warnf(temp, args...)
}

func (applg *AppLogger) Errorf(temp string, args ...interface{}) {
	applg.slg.Errorf(temp, args...)
}

// Fatalf logs a message at FatalLevel with SugaredLogger
// The logger then calls os.Exit(1), even if logging at FatalLevel is disabled
func (applg *AppLogger) Fatalf(temp string, args ...interface{}) {
	applg.slg.Fatalf(temp, args...)
}

type Option func(conf zap.Config)

// SetDebug 用于创建logger时,设置是否处于开发环境
func SetDebug(debug bool) Option {
	encoding := "console"
	if debug {
		encoding = "json"
	}
	return func(conf zap.Config) {
		conf.Development = debug
		conf.Encoding = encoding
	}
}

// SetLevel 创建Logger实例时, 设置日志等级
func SetLevel(lv Level) Option {
	return func(conf zap.Config) {
		conf.Level = zap.NewAtomicLevelAt(zapcore.Level(lv))
	}
}

// NewLogger 创建logger实例对象:
// opts: logger包提供的对logger配置参数的配置方法
func NewLogger(opts ...Option) (Logger, error) {
	if Vicglog != nil {
		return Vicglog, nil
	}
	return createGlobalLogger(opts...)
}

// createGlobalLogger 创建全局唯一的logger对象.
func createGlobalLogger(opts ...Option) (*AppLogger, error) {
	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(zapcore.InfoLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "console",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "T",
			LevelKey:       "L",
			NameKey:        "N",
			CallerKey:      "C",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "M",
			StacktraceKey:  "S",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,             // 等级显示标识
			EncodeTime:     zapcore.TimeEncoderOfLayout(timeFormat), // 时间格式
			EncodeDuration: zapcore.StringDurationEncoder,           // 调用时间序列化
			EncodeCaller:   zapcore.ShortCallerEncoder,              // 调用文件路径
		},
		OutputPaths:      otputPaths,
		ErrorOutputPaths: otputPaths,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	lg, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}
	return &AppLogger{cfg: &cfg, slg: lg.Sugar(), lg: lg, level: cfg.Level}, nil
}
