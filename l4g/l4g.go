package l4g

import (
	"fmt"
	l4g "github.com/zl03jsj/log4go"
	"gitlab.forceup.in/zengliang/rpc2-center/utils"
	"os"
	"path"
	"strings"
	"time"
)

var (
	lvl          l4g.Level
	is_writefile = true
	filefmt      = "rawtext"
	log_dir      string
	fmt_time     = "[%T] [%L] (%s) %M"

	def_logfile    = "default"
	def_loger_name = "default"

	all_logers map[string]*ILoger
)

func init() {
	if err := Initialize("./data/logs/", "all", "default.log", "rawtext", true); err != nil {
		panic(err)
	}
}

func Initialize(logdir, loglvl, file, _filefmt string, iswritefile bool) error {
	if !utils.Isdirectoryexist(logdir) {
		if err := os.MkdirAll(logdir, os.ModePerm); err != nil {
			return fmt.Errorf("directory(%s) doesn't exist, mkdir faild with error:%s",
				logdir, err.Error())
		}
	}

	log_dir = logdir
	def_logfile = file
	is_writefile = iswritefile

	if _filefmt == "json" || _filefmt == "rawtext" {
		filefmt = _filefmt
	}

	all_logers = make(map[string]*ILoger)

	SetLvl(loglvl)

	BuildL4g(def_loger_name, def_logfile)

	return nil
}

func Default() *ILoger {
	return GetL4g(def_loger_name)
}

func GetL4g(name string) *ILoger {
	l, isok := all_logers[name]

	if !isok || nil == l {
		l = all_logers[def_loger_name]
		l.Warn("there are no l4g named '%s', return '%s' ", name, def_loger_name)
	}

	return l
}

func SetIsWriteLogfile(islogfile bool) {
	is_writefile = islogfile
}

func BuildL4g(name string, filename string) *ILoger {
	name = strings.Trim(name, " ")
	filename = strings.Trim(filename, " ")

	if all_logers == nil {
		return nil
	}

	if l, isok := all_logers[name]; isok {
		return l
	}
	if name == "" {
		return nil
	}

	if filename == "" {
		filename = name
	}

	if l := len(filename); l < 4 || (l >= 4 && filename[l-4:] != ".log") {
		filename += "-%P-%T.log"
	} else {
		filename = filename[:l-4] + "-%P-%T.log"
	}

	l := make(l4g.Logger)

	maxsize := 20 * 1024 * 1024
	maxline := 100000

	if is_writefile {
		filename = path.Join(log_dir, filename)
		if filefmt == "json" {
			flw := l4g.NewJSONLogWriter(filename, true)
			flw.SetFormat("%A %B " + fmt_time)
			flw.SetRotateSize(maxsize)
			flw.SetRotateLines(maxline)
			flw.SetRotateDaily(true)
			flw.SetModuleName(name)
			l.AddFilter("file", lvl, flw)
		} else { // filefmt == "rawtext"
			flw := l4g.NewFileLogWriter(filename, true)
			flw.SetFormat(fmt_time)
			flw.SetRotateSize(maxsize)
			flw.SetRotateLines(maxline)
			flw.SetRotateDaily(true)
			l.AddFilter("file", lvl, flw)
		}
	}

	consolWriter := l4g.NewConsoleLogWriter()
	consolWriter.SetFormat(fmt_time)
	l.AddFilter("stdout", lvl, consolWriter)

	all_logers[name] = &ILoger{l}

	return &ILoger{l}
}

func SetLvl(lvlstring string) {
	lvl = l4g.LvlFromString(strings.ToUpper(lvlstring))

	l4g.Global.SetFilterLvl("all", lvl)

	for _, lg := range all_logers {
		lg.SetFilterLvl("all", lvl)
	}
}

func Close(name string) {
	if name == "all" {
		for _, l := range all_logers {
			l.Close()
			delete(all_logers, name)
		}
	} else {
		if l, exist := all_logers[name]; exist {
			l.Close()
			delete(all_logers, name)
		}
	}
}

type ILoger struct{ l4g.Logger }

func (self *ILoger) Debug(arg0 string, args ...interface{}) {
	self.Logger.DebugSkip(2, arg0, args...)
}
func (self *ILoger) Info(arg0 string, args ...interface{}) {
	self.Logger.InfoSkip(2, arg0, args...)
}
func (self *ILoger) Trace(arg0 string, args ...interface{}) {
	self.Logger.TraceSkip(2, arg0, args...)
}
func (self *ILoger) Warns(arg0 string, args ...interface{}) {
	self.Logger.WarnSkip(2, arg0, args...)
}
func (self *ILoger) Error(arg0 string, args ...interface{}) {
	self.Logger.ErrorSkip(2, arg0, args...)
}
func (self *ILoger) Fatal(arg0 string, args ...interface{}) {
	var err = self.Logger.ErrorSkip(2, arg0, args...)
	time.Sleep(time.Second * 2)
	panic(err)
}
