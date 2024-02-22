// Copyright (c) 2023, The OTNS Authors.
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
	"os"
	"runtime/debug"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

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

var (
	cfg             zap.Config
	zaplogger       *zap.Logger
	currentLevel    Level
	isLogToTerminal bool
	cbStdout        StdoutCallback
	zapLevels       = []zapcore.Level{zapcore.FatalLevel + 1, zapcore.FatalLevel, zapcore.PanicLevel,
		zapcore.ErrorLevel, zapcore.WarnLevel, zapcore.InfoLevel, zapcore.InfoLevel, zapcore.DebugLevel,
		zapcore.DebugLevel, zapcore.DebugLevel}
)

func init() {
	o, _ := os.Stdout.Stat()
	if (o.Mode() & os.ModeCharDevice) == os.ModeCharDevice {
		isLogToTerminal = true
	}

	var err error
	cfgJson := []byte(`{
		"level": "debug",
	"outputPaths": ["stderr"],
	"errorOutputPaths": ["stderr"],
	"encoding": "console",
		"encoderConfig": {
		"messageKey": "message",
			"levelKey": "level",
			"levelEncoder": "lowercase"
	}
}`)
	currentLevel = DefaultLevel

	if err = json.Unmarshal(cfgJson, &cfg); err != nil {
		panic(err)
	}
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	rebuildLoggerFromCfg()
}

// SetLevel sets the log level
func SetLevel(lv Level) {
	currentLevel = lv
}

// GetLevel get the current log level
func GetLevel() Level {
	return currentLevel
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

// SetOutput sets the output writer
// e.g. logger.SetOutput([]string{"stderr", "otns.log"}) // for @DEBUG: generate a log output file.
func SetOutput(outputs []string) {
	cfg.OutputPaths = outputs
	rebuildLoggerFromCfg()
}

func rebuildLoggerFromCfg() {
	if newLogger, err := cfg.Build(); err == nil {
		if zaplogger != nil {
			_ = zaplogger.Sync()
		}
		zaplogger = newLogger
	} else {
		panic(err)
	}
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
	timeStr := time.Now().Format("2006-01-02 15:04:05.000") + " - "
	zaplogger.Log(zapLevels[level-MinLevel], timeStr+getMessage(format, args))
	if isLogToTerminal && cbStdout != nil {
		cbStdout.OnStdout()
	}
}

// logAlways is a helper func that doesn't check level prior to logging to zaplogger.
func logAlways(level Level, msg string) {
	if isLogToTerminal {
		_, _ = fmt.Fprint(os.Stdout, "\033[2K\r") // ANSI sequence to clear the CLI line
	}
	timeStr := time.Now().Format("2006-01-02 15:04:05.000") + " - "
	zaplogger.Log(zapLevels[level-MinLevel], timeStr+msg)
	if isLogToTerminal && cbStdout != nil {
		cbStdout.OnStdout()
	}
}

// Println prints a message for the user at the current console/CLI, to stdout, without logging fields.
func Println(msg string) {
	if isLogToTerminal {
		_, _ = fmt.Fprint(os.Stdout, "\033[2K\r") // ANSI sequence to clear the CLI line
	}
	_, _ = fmt.Fprint(os.Stdout, msg+"\n")
	if isLogToTerminal && cbStdout != nil {
		cbStdout.OnStdout()
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
