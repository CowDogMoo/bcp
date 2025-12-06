/*
Copyright © 2025 Jayson Grace <jayson.e.grace@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package logging

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

type Logger struct {
	level         Level
	format        string
	consoleWriter io.Writer
}

var defaultLogger *Logger

func init() {
	defaultLogger = &Logger{
		level:         InfoLevel,
		format:        "text",
		consoleWriter: os.Stderr,
	}
}

func Init(format string, level string) {
	defaultLogger.format = format
	defaultLogger.level = parseLevel(level)
}

func parseLevel(level string) Level {
	switch strings.ToLower(level) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		return InfoLevel
	}
}

func (l *Logger) shouldLog(level Level) bool {
	return level >= l.level
}

func (l *Logger) formatMessage(level Level, format string, args ...interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)

	if l.format == "json" {
		return fmt.Sprintf(`{"time":"%s","level":"%s","message":"%s"}`,
			timestamp, level.String(), message)
	}

	return fmt.Sprintf("[%s] %s: %s", timestamp, level.String(), message)
}

func Debug(format string, args ...interface{}) {
	if defaultLogger.shouldLog(DebugLevel) {
		msg := defaultLogger.formatMessage(DebugLevel, format, args...)
		fmt.Fprintln(defaultLogger.consoleWriter, msg)
	}
}

func Info(format string, args ...interface{}) {
	if defaultLogger.shouldLog(InfoLevel) {
		msg := defaultLogger.formatMessage(InfoLevel, format, args...)
		fmt.Fprintln(defaultLogger.consoleWriter, msg)
	}
}

func Warn(format string, args ...interface{}) {
	if defaultLogger.shouldLog(WarnLevel) {
		msg := defaultLogger.formatMessage(WarnLevel, format, args...)
		fmt.Fprintln(defaultLogger.consoleWriter, msg)
	}
}

func Error(format string, args ...interface{}) {
	if defaultLogger.shouldLog(ErrorLevel) {
		msg := defaultLogger.formatMessage(ErrorLevel, format, args...)
		fmt.Fprintln(defaultLogger.consoleWriter, msg)
	}
}
