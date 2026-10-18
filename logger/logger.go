package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"self_medie_script_tool/config"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	ZapLogger *zap.Logger
	Sugar     *zap.SugaredLogger
)

// Init 根据config里的log配置初始化日志
func Init(logCfg config.LogConfig) {
	// 解析日志级别
	var lvl zapcore.Level
	_ = lvl.UnmarshalText([]byte(logCfg.Level))

	// 自动创建日志目录
	if logCfg.File != "" {
		logDir := filepath.Dir(logCfg.File)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			fmt.Printf("警告：创建日志目录失败，不输出文件日志，err=%v\n", err)
			logCfg.File = ""
		}
	}

	// 编码器：dev用控制台彩色，生产用json
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)

	var cores []zapcore.Core

	// 如果配置了日志文件，增加文件输出core，带lumberjack滚动
	if logCfg.File != "" {
		lumberJackLogger := &lumberjack.Logger{
			Filename:   logCfg.File,
			MaxSize:    logCfg.MaxSize,
			MaxBackups: logCfg.MaxBackups,
			MaxAge:     logCfg.MaxAge,
			Compress:   logCfg.Compress,
		}
		fileWriter := zapcore.AddSync(lumberJackLogger)
		fileCore := zapcore.NewCore(jsonEncoder, zapcore.Lock(fileWriter), lvl)
		cores = append(cores, fileCore)
	}

	// 兜底：如果cores为空（日志文件配置为空或者目录创建失败），避免zap panic
	if len(cores) == 0 {
		// 极端兜底，完全无输出，也可以选择强制输出控制台做应急
		ZapLogger = zap.NewNop()
		Sugar = ZapLogger.Sugar()
		return
	}

	// 合并多个core
	core := zapcore.NewTee(cores...)
	ZapLogger = zap.New(core, zap.AddCaller()) // AddCaller 打印代码文件+行号
	Sugar = ZapLogger.Sugar()
}

// Sync 刷写缓冲区，程序退出时defer调用
func Sync() {
	_ = ZapLogger.Sync()
}
