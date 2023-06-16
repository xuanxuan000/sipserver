package utils

import (
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var aclog *zap.SugaredLogger

var atom zap.AtomicLevel

// LoggerInit 这个初始化，在程序在云下主机上运行时使用
func LoggerInit() (*zap.SugaredLogger, zap.AtomicLevel) {

	fileName := "/var/log/" + filepath.Base(os.Args[0]) + ".log"
	atom = zap.NewAtomicLevel()
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    100,  // megabytes 日志超过指定大小就自动分割,分割之前的日志就变为老日志
		MaxBackups: 10,   //最多备份的老的日志文件个数
		MaxAge:     280,  // days
		Compress:   true, //日志压缩
		LocalTime:  true,
	})

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		// zapcore.NewJSONEncoder(encoderCfg),
		w,
		atom,
	)
	aclog = zap.New(core).Sugar()
	//zap.AddCaller() 用来添加caller信息 如 "caller":"tool/zaplogger.go:87"
	// aclog = zap.New(core, zap.AddCaller()).Sugar()

	atom.SetLevel(zap.DebugLevel)
	go SignalHandle()
	return aclog, atom
}

// LoggerToStdoutInit 这个初始化，在程序在本地运行时使用
func LoggerToStdoutInit() (*zap.SugaredLogger, zap.AtomicLevel) {

	atom = zap.NewAtomicLevel()
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		// zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		atom,
	)
	aclog = zap.New(core).Sugar()
	//zap.AddCaller() 用来添加caller信息 如 "caller":"tool/zaplogger.go:87"
	// aclog = zap.New(core, zap.AddCaller()).Sugar()

	atom.SetLevel(zap.DebugLevel)
	go SignalHandle()
	return aclog, atom
}

// SetLogger
func SetLogger(acLogger *zap.SugaredLogger, atomLever zap.AtomicLevel) {
	aclog = acLogger
	atom = atomLever
}

func fileInfo() string {

	pc, file, line, _ := runtime.Caller(2)
	short := file

	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			short = file[i+1:]
			break
		}
	}
	f := runtime.FuncForPC(pc)
	fn := f.Name()

	for i := len(fn) - 1; i > 0; i-- {
		if fn[i] == '.' {
			fn = fn[i+1:]
			break
		}
	}

	return short + ":" + strconv.Itoa(line) + "[" + fn + "]"

}

func SetLogLevel(logLevel string) {
	switch logLevel {
	case "DEBUG":
		atom.SetLevel(zap.DebugLevel)
	case "WARNING":
		atom.SetLevel(zap.WarnLevel)
	case "ERROR":
		atom.SetLevel(zap.ErrorLevel)
	case "INFO":
		atom.SetLevel(zap.InfoLevel)
	}
}

func Debug(v ...interface{}) {
	aclog.Named(fileInfo()).Debug(v...)

}
func Debugf(format string, v ...interface{}) {
	aclog.Named(fileInfo()).Debugf(format, v...)
}

func Error(v ...interface{}) {
	aclog.Named(fileInfo()).Error(v...)
}
func Errorf(format string, v ...interface{}) {
	aclog.Named(fileInfo()).Errorf(format, v...)
}

func Warning(v ...interface{}) {
	aclog.Named(fileInfo()).Warn(v...)
}
func Warningf(format string, v ...interface{}) {
	aclog.Named(fileInfo()).Warnf(format, v...)
}

func Info(v ...interface{}) {
	aclog.Named(fileInfo()).Info(v...)
}
func Infof(format string, v ...interface{}) {
	aclog.Named(fileInfo()).Infof(format, v...)
}

func SignalHandle() {
	Info("Start SignalHandle")
	t1 := time.NewTimer(1 * time.Hour)
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGUSR1, syscall.SIGUSR2)
	for {
		select {
		case sig := <-ch:
			switch sig {
			case syscall.SIGUSR1:
				SetLogLevel("DEBUG")
				t1.Reset(1 * time.Hour)
			case syscall.SIGUSR2:
				SetLogLevel("ERROR")
				t1.Reset(1 * time.Hour)
				t1.Stop()
			default:
			}
		case <-t1.C:
			SetLogLevel("ERROR")
			t1.Stop()
		}
	}
}
