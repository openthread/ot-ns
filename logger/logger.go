// Copyright (c) 2023-2026, The OTNS Authors.
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
// 3. Neither the name of the copyright holder nor the
//    names of its contributors may be used to endorse or promote products
//    derived from this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/term"

	. "github.com/openthread/ot-ns/types"
)

// Level is the log-level for logging what happens in the simulation as a whole, or to watch an
// individual node. Values inherit OT logging.h values and extend these with OT-NS specific items.
type Level int8

const (
	MicroLevel   Level = 7
	TraceLevel   Level = 6
	DebugLevel   Level = 5
	InfoLevel    Level = 4
	NoteLevel    Level = 3
	WarnLevel    Level = 2
	ErrorLevel   Level = 1
	PanicLevel   Level = 0
	FatalLevel   Level = -1
	OffLevel     Level = -2
	MinLevel           = OffLevel
	DefaultLevel       = InfoLevel
)

type logEntry struct {
	NodeId NodeId
	Level  Level
	Msg    string
}

type StdoutCallback interface {
	OnStdout()
}

const (
	dateTimeFormat = "2006-01-02 15:04:05.000"
)

var (
	cfg               zap.Config
	zaplogger         *zap.Logger
	currentLevel      Level          = DefaultLevel
	isLogToTerminal                  = true
	cbStdout          StdoutCallback = nil
	logFileHandle     *os.File       = nil
	logPath                          = ""
	logFileWriterInst                = &logFileWriter{}
	zapLevels                        = []zapcore.Level{zapcore.FatalLevel + 1, zapcore.FatalLevel, zapcore.PanicLevel,
		zapcore.ErrorLevel, zapcore.WarnLevel, zapcore.InfoLevel, zapcore.InfoLevel, zapcore.DebugLevel,
		zapcore.DebugLevel, zapcore.DebugLevel}
)

func init() {
	cfgJson := []byte(`{
        "level": "debug",
        "outputPaths": ["stderr"],
        "errorOutputPaths": ["stderr"],
        "encoding": "console",
        "encoderConfig": {
            "messageKey": "message",
            "levelKey": "level",
            "levelEncoder": "lowercase",
            "timeKey": "timestamp",
			"timeEncoder": "iso8601"
        }
    }`)

	if err := json.Unmarshal(cfgJson, &cfg); err != nil {
		panic(err)
	}

	cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(dateTimeFormat))
	}

	rebuildLoggerFromCfg()
}

func Init(logToStdout bool, logToFile bool, logFileName string, simId int) {
	var err error
	logPath = logFileName

	if logToFile {
		// Open the log file here (not inside Zap) so we can store the file handle in our package.
		// os.O_APPEND is used to enable multiple goroutines to write to the same handle.
		// os.O_TRUNC ensures any prior log file from a previous run is overwritten cleanly.
		logFileHandle, err = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			Errorf("Error: failed to open log file: %v\n", err)
		} else {
			header := fmt.Sprintf("#\n# OTNS log for sim-ID %d created %s\n#\n", simId,
				time.Now().Format(time.RFC3339))
			_, err = logFileHandle.WriteString(header)
			if err != nil {
				Errorf("Error: failed to write log file header: %v\n", err)
				_ = logFileHandle.Close()
				logFileHandle = nil
			}
		}
	}

	isLogToTerminal = logToStdout && term.IsTerminal(int(os.Stdout.Fd()))

	rebuildLoggerFromCfg()
}

// SetLevel sets the log level
func SetLevel(lv Level) {
	if currentLevel != lv {
		currentLevel = lv
		Println(fmt.Sprintf("%s         Log level changed to %s", time.Now().Format(dateTimeFormat), GetLevelString(lv)), false, true)
	}
}

// GetLevel get the current log level
func GetLevel() Level {
	return currentLevel
}

// GetLogWriter returns an io.Writer that writes directly to the log file, or to /dev/null in case
// there is no log file active.
func GetLogWriter() io.Writer {
	return logFileWriterInst
}

type logFileWriter struct{}

func (w *logFileWriter) Write(p []byte) (n int, err error) {
	if logFileHandle != nil {
		return logFileHandle.Write(p)
	}
	return len(p), nil
}

// SetStdoutCallback sets a callback, that the logger will call when new log content was written to stdout/stderr.
func SetStdoutCallback(cb StdoutCallback) {
	cbStdout = cb
}

// TraceError prints the stack and error
func TraceError(format string, args ...interface{}) {
	Error(string(debug.Stack()))
	Errorf(format, args...)
}

func rebuildLoggerFromCfg() {
	if zaplogger != nil {
		_ = zaplogger.Sync()
	}

	encoder := zapcore.NewConsoleEncoder(cfg.EncoderConfig)
	// Accept all levels — Go-level checks in Logf/logAlways gate what actually reaches zap.
	allLevels := zap.LevelEnablerFunc(func(zapcore.Level) bool { return true })

	var core zapcore.Core
	switch {
	case logFileHandle != nil && isLogToTerminal:
		core = zapcore.NewTee(
			zapcore.NewCore(encoder, zapcore.AddSync(os.Stderr), allLevels),
			zapcore.NewCore(encoder, zapcore.AddSync(logFileHandle), allLevels),
		)
	case logFileHandle != nil:
		core = zapcore.NewCore(encoder, zapcore.AddSync(logFileHandle), allLevels)
	default:
		core = zapcore.NewCore(encoder, zapcore.AddSync(os.Stderr), allLevels)
	}

	zaplogger = zap.New(core)
}

// getMessage formats a string efficiently with Sprint, Sprintf, or neither.
func getMessage(template string, fmtArgs []interface{}) string {
	if len(fmtArgs) == 0 {
		return template
	}

	if template != "" {
		return fmt.Sprintf(template, fmtArgs...)
	}

	if len(fmtArgs) == 1 {
		if str, ok := fmtArgs[0].(string); ok {
			return str
		}
	}
	return fmt.Sprint(fmtArgs...)
}

// Log outputs the log message/object at specified level using logger.
func Log(level Level, msg interface{}) {
	if level > currentLevel {
		return
	}
	Logf(level, "", []interface{}{msg})
}

// Logf outputs formatted log message at specified level using logger.
func Logf(level Level, format string, args []interface{}) {
	if level > currentLevel {
		return
	}
	if isLogToTerminal {
		_, _ = fmt.Fprint(os.Stdout, "\033[2K\r") // ANSI sequence to clear the CLI line
	}
	zaplogger.Log(zapLevels[level-MinLevel], getMessage(format, args))
	if isLogToTerminal && cbStdout != nil {
		cbStdout.OnStdout()
	}
}

// logAlways is a helper func that doesn't check level and always logs to zaplogger.
func logAlways(level Level, msg string) {
	if isLogToTerminal {
		_, _ = fmt.Fprint(os.Stdout, "\033[2K\r") // ANSI sequence to clear the CLI line
	}
	zaplogger.Log(zapLevels[level-MinLevel], msg)
	if isLogToTerminal && cbStdout != nil {
		cbStdout.OnStdout()
	}
}

// Println prints a message to console and/or log file, without using any log line formatting.
func Println(msg string, toConsole bool, toLogFile bool) {
	if toConsole {
		if isLogToTerminal {
			_, _ = fmt.Fprint(os.Stdout, "\033[2K\r") // ANSI sequence to clear the CLI line
		}
		_, _ = fmt.Fprintln(os.Stdout, msg)
		if isLogToTerminal && cbStdout != nil {
			cbStdout.OnStdout()
		}
	}
	if toLogFile && logFileHandle != nil {
		_, _ = logFileHandle.WriteString(msg + "\n")
	}
}

func Tracef(format string, args ...interface{}) {
	Logf(TraceLevel, format, args)
}

func Debugf(format string, args ...interface{}) {
	Logf(DebugLevel, format, args)
}

func Infof(format string, args ...interface{}) {
	Logf(InfoLevel, format, args)
}

func Warnf(format string, args ...interface{}) {
	Logf(WarnLevel, format, args)
}

func Errorf(format string, args ...interface{}) {
	Logf(ErrorLevel, format, args)
}

func Panicf(format string, args ...interface{}) {
	Logf(PanicLevel, format, args)
}

func Fatalf(format string, args ...interface{}) {
	Logf(FatalLevel, format, args)
}

func Error(args ...interface{}) {
	Log(ErrorLevel, args)
}

func Panic(args ...interface{}) {
	Log(PanicLevel, args)
}

func Fatal(args ...interface{}) {
	Log(FatalLevel, args)
}

func PanicIfError(err error, args ...interface{}) {
	if len(args) == 0 {
		args = []interface{}{err}
	}
	if err != nil {
		Panic(args...)
	}
}

func PanicfIfError(err error, format string, args ...interface{}) {
	if err != nil {
		Panicf(format, args...)
	}
}

func FatalIfError(err error, args ...interface{}) {
	if len(args) == 0 {
		args = []interface{}{err}
	}
	if err != nil {
		Fatal(args...)
	}
}

func FatalfIfError(err error, format string, args ...interface{}) {
	if err != nil {
		Fatalf(format, args...)
	}
}

type assertLogger struct{}

func (t assertLogger) Errorf(format string, args ...interface{}) {
	Panicf(format, args...)
}

func AssertEqual(expected, actual interface{}, msgAndArgs ...interface{}) bool {
	return assert.Equal(assertLogger{}, expected, actual, msgAndArgs...)
}

func AssertEqualf(expected interface{}, actual interface{}, msg string, args ...interface{}) bool {
	return assert.Equalf(assertLogger{}, expected, actual, msg, args...)
}

func AssertNil(object interface{}, msgAndArgs ...interface{}) bool {
	return assert.Nil(assertLogger{}, object, msgAndArgs...)
}

func AssertNotNil(object interface{}, msgAndArgs ...interface{}) bool {
	return assert.NotNil(assertLogger{}, object, msgAndArgs...)
}

func AssertNilF(object interface{}, msg string, args ...interface{}) bool {
	return assert.Nilf(assertLogger{}, object, msg, args...)
}

func AssertNotNilF(object interface{}, msg string, args ...interface{}) bool {
	return assert.NotNilf(assertLogger{}, object, msg, args...)
}

func AssertTrue(value bool, msgAndArgs ...interface{}) bool {
	return assert.True(assertLogger{}, value, msgAndArgs...)
}

func AssertFalse(value bool, msgAndArgs ...interface{}) bool {
	return assert.False(assertLogger{}, value, msgAndArgs...)
}

func AssertTruef(value bool, msg string, args ...interface{}) bool {
	return assert.Truef(assertLogger{}, value, msg, args...)
}

func AssertFalsef(value bool, msg string, args ...interface{}) bool {
	return assert.Falsef(assertLogger{}, value, msg, args...)
}
