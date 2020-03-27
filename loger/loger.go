package loger

import (
	"fmt"
)

type ILoger interface {
	Debug(fmt_str string, args ...interface{})
	Info(fmt_str string, args ...interface{})
	Trace(fmt_str string, args ...interface{})
	Warns(fmt_str string, args ...interface{})
	Error(fmt_str string, args ...interface{})
	Fatal(fmt_str string, args ...interface{})
}

type MyLoger struct{}

func (self *MyLoger) printf(fmts string, args ...interface{}) {
	fmt.Printf(fmts, args...)
}
func (self *MyLoger) Debug(fmt_str string, args ...interface{}) {
	self.printf(fmt_str, args...)
}
func (self *MyLoger) Info(fmt_str string, args ...interface{}) {
	self.printf(fmt_str, args...)
}
func (self *MyLoger) Trace(fmt_str string, args ...interface{}) {
	self.printf(fmt_str, args...)
}
func (self *MyLoger) Warns(fmt_str string, args ...interface{}) {
	self.printf(fmt_str, args...)
}
func (self *MyLoger) Error(fmt_str string, args ...interface{}) {
	self.printf(fmt_str, args...)
}
func (self *MyLoger) Fatal(fmt_str string, args ...interface{}) {
	panic(fmt.Sprintf(fmt_str, args...))
}
