package logger

import "log"

type Logger interface {
	Debug(arg0 string, args ...interface{})
	Info(arg0 string, args ...interface{})
	Trace(arg0 string, args ...interface{})
	Warns(arg0 string, args ...interface{})
	Error(arg0 string, args ...interface{})
	Fatal(arg0 string, args ...interface{})
}

var (
	GlobalLogger Logger
)

func InitLogger(logger Logger)  {
	GlobalLogger = logger
	if GlobalLogger == nil {
		log.Fatalln("logger is nil")
	}
}

type MyLogger struct {
}

func (l *MyLogger) Debug(arg0 string, args ...interface{}) {
	log.Printf(arg0, args...)
}
func (l *MyLogger) Info(arg0 string, args ...interface{}) {
	log.Printf(arg0, args...)
}

func (l *MyLogger) Trace(arg0 string, args ...interface{}) {
	log.Printf(arg0, args...)
}

func (l *MyLogger) Warns(arg0 string, args ...interface{}) {
	log.Printf(arg0, args...)
}
func (l *MyLogger) Error(arg0 string, args ...interface{}) {
	log.Printf(arg0, args...)
}
func (l *MyLogger) Fatal(arg0 string, args ...interface{}) {
	log.Fatalf(arg0, args...)
}

