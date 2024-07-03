package logger

import (
	"fmt"
	"gedis/util"
	"log"
	"os"
	"path"
	"runtime"
	"sync"
	"time"
)

const (
	maxLogMessageNum = 1e5
	callerDepth      = 2

	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
)

type LogLevel int

const (
	NULL LogLevel = iota

	FATAL
	ERROR
	WARN
	INFO
	DEBUG
)

type logMessage struct {
	level LogLevel
	msg   string
}

func (m *logMessage) reset() {
	m.level = NULL
	m.msg = ""
}

var levelFlags = []string{"", "Fatal", "Error", "Warn", " Info", "Debug"}

type logger struct {
	logFile    *os.File
	logStd     *log.Logger
	logMsgChan chan *logMessage
	logMsgPool *sync.Pool
	logLevel   LogLevel
	close      chan struct{}
}

func (l *logger) Close() {
	close(l.close)
}

func (l *logger) writeLog(level LogLevel, callerDepth int, msg string) {
	var formattedMsg string
	_, file, line, ok := runtime.Caller(callerDepth)
	if ok {
		formattedMsg = fmt.Sprintf("[%s] [%s:%d] %s", levelFlags[level], file, line, msg)
	} else {
		formattedMsg = fmt.Sprintf("[%s] %s", levelFlags[level], msg)
	}

	logMsg := l.logMsgPool.Get().(*logMessage)
	logMsg.level = level
	logMsg.msg = formattedMsg
	l.logMsgChan <- logMsg
}

var defaultLogger = newStdLogger()

func newStdLogger() *logger {
	stdLogger := &logger{
		logStd:     log.New(os.Stdout, "", log.LstdFlags),
		logMsgChan: make(chan *logMessage, maxLogMessageNum),
		logLevel:   DEBUG,
		close:      make(chan struct{}),
		logMsgPool: &sync.Pool{
			New: func() any {
				return &logMessage{}
			},
		},
	}

	go func() {
		for {
			select {
			case <-stdLogger.close:
				return
			case logMsg := <-stdLogger.logMsgChan:
				msg := logMsg.msg
				switch logMsg.level {
				case DEBUG:
					msg = Blue + msg + Reset
				case INFO:
					msg = Green + msg + Reset
				case WARN:
					msg = Yellow + msg + Reset
				case ERROR, FATAL:
					msg = Red + msg + Red
				}
				stdLogger.logStd.Output(0, msg)
				logMsg.reset()
				stdLogger.logMsgPool.Put(logMsg)
			}
		}
	}()

	return stdLogger
}

type Settings struct {
	Path       string `yaml:"path"`
	Name       string `yaml:"name"`
	Ext        string `yaml:"ext"`
	DateFormat string `yaml:"dateFormat"`
}

func newFileLogger(setting *Settings) (*logger, error) {
	filename := fmt.Sprintf("%s-%s.%s", setting.Name, time.Now().Format(setting.DateFormat), setting.Ext)
	fd, err := util.OpenFile(filename, setting.Path)
	if err != nil {
		return nil, err
	}
	fileLogger := &logger{
		logFile:    fd,
		logStd:     log.New(os.Stdout, "", log.LstdFlags),
		logMsgChan: make(chan *logMessage, maxLogMessageNum),
		logLevel:   DEBUG,
		close:      make(chan struct{}),
		logMsgPool: &sync.Pool{
			New: func() any {
				return &logMessage{}
			},
		},
	}

	go func() {
		for {
			select {
			case <-fileLogger.close:
				return
			case logMsg := <-fileLogger.logMsgChan:
				logFilename := fmt.Sprintf("%s-%s.%s", setting.Name, time.Now().Format(setting.DateFormat), setting.Ext)
				if path.Join(setting.Path, logFilename) != fileLogger.logFile.Name() {
					fd, err := util.OpenFile(logFilename, setting.Path)
					if err != nil {
						panic("创建日志文件失败")
					}
					fileLogger.logFile.Close()
					fileLogger.logFile = fd
				}
				msg := logMsg.msg
				switch logMsg.level {
				case DEBUG:
					msg = Blue + msg + Reset
				case INFO:
					msg = Green + msg + Reset
				case WARN:
					msg = Yellow + msg + Reset
				case ERROR, FATAL:
					msg = Red + msg + Red
				}
				fileLogger.logStd.Output(0, msg)
				fileLogger.logFile.WriteString(time.Now().Format(util.DateFormat) + " " + logMsg.msg + util.CRLF)
				logMsg.reset()
				fileLogger.logMsgPool.Put(logMsg)
			}
		}
	}()
	return fileLogger, nil
}

func SetUp(settings *Settings) {
	defaultLogger.Close()
	logger, err := newFileLogger(settings)
	if err != nil {
		panic(err)
	}
	defaultLogger = logger
}

func SetLoggerLevel(logLevel LogLevel) {
	defaultLogger.logLevel = logLevel
}

// #######打印相关的函数#############//
func Debug(v ...any) {
	if defaultLogger.logLevel >= DEBUG {
		msg := fmt.Sprint(v...)
		defaultLogger.writeLog(DEBUG, callerDepth, msg)
	}
}

func Debugf(format string, v ...any) {
	if defaultLogger.logLevel >= DEBUG {
		msg := fmt.Sprintf(format, v...)
		defaultLogger.writeLog(DEBUG, callerDepth, msg)
	}
}

func Info(v ...any) {
	if defaultLogger.logLevel >= INFO {
		msg := fmt.Sprint(v...)
		defaultLogger.writeLog(INFO, callerDepth, msg)
	}
}

func Infof(format string, v ...any) {
	if defaultLogger.logLevel >= INFO {
		msg := fmt.Sprintf(format, v...)
		defaultLogger.writeLog(INFO, callerDepth, msg)
	}
}

func Warn(v ...any) {
	if defaultLogger.logLevel >= WARN {
		msg := fmt.Sprint(v...)
		defaultLogger.writeLog(WARN, callerDepth, msg)
	}
}

func Warnf(format string, v ...any) {
	if defaultLogger.logLevel >= WARN {
		msg := fmt.Sprintf(format, v...)
		defaultLogger.writeLog(WARN, callerDepth, msg)
	}
}

func Error(v ...any) {
	if defaultLogger.logLevel >= ERROR {
		msg := fmt.Sprint(v...)
		defaultLogger.writeLog(ERROR, callerDepth, msg)
	}
}

func Errorf(format string, v ...any) {
	if defaultLogger.logLevel >= ERROR {
		msg := fmt.Sprintf(format, v...)
		defaultLogger.writeLog(ERROR, callerDepth, msg)
	}
}

func Fatal(v ...any) {
	if defaultLogger.logLevel >= FATAL {
		msg := fmt.Sprint(v...)
		defaultLogger.writeLog(FATAL, callerDepth, msg)
	}
}

func Fatalf(format string, v ...any) {
	if defaultLogger.logLevel >= FATAL {
		msg := fmt.Sprintf(format, v...)
		defaultLogger.writeLog(FATAL, callerDepth, msg)
	}
}
