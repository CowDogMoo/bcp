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
	"bytes"
	"strings"
	"testing"
)

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{WarnLevel, "WARN"},
		{ErrorLevel, "ERROR"},
		{Level(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("Level.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
	}{
		{"debug", DebugLevel},
		{"DEBUG", DebugLevel},
		{"info", InfoLevel},
		{"INFO", InfoLevel},
		{"warn", WarnLevel},
		{"warning", WarnLevel},
		{"WARN", WarnLevel},
		{"error", ErrorLevel},
		{"ERROR", ErrorLevel},
		{"unknown", InfoLevel}, // defaults to InfoLevel
		{"", InfoLevel},        // defaults to InfoLevel
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := parseLevel(tt.input); got != tt.expected {
				t.Errorf("parseLevel(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestInit(t *testing.T) {
	tests := []struct {
		name          string
		format        string
		level         string
		expectedLevel Level
	}{
		{"text format debug level", "text", "debug", DebugLevel},
		{"json format info level", "json", "info", InfoLevel},
		{"text format warn level", "text", "warn", WarnLevel},
		{"json format error level", "json", "error", ErrorLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init(tt.format, tt.level)

			if defaultLogger.format != tt.format {
				t.Errorf("Init() format = %v, want %v", defaultLogger.format, tt.format)
			}
			if defaultLogger.level != tt.expectedLevel {
				t.Errorf("Init() level = %v, want %v", defaultLogger.level, tt.expectedLevel)
			}
		})
	}
}

func TestShouldLog(t *testing.T) {
	logger := &Logger{
		level: WarnLevel,
	}

	tests := []struct {
		name     string
		level    Level
		expected bool
	}{
		{"debug should not log", DebugLevel, false},
		{"info should not log", InfoLevel, false},
		{"warn should log", WarnLevel, true},
		{"error should log", ErrorLevel, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := logger.shouldLog(tt.level); got != tt.expected {
				t.Errorf("shouldLog(%v) = %v, want %v", tt.level, got, tt.expected)
			}
		})
	}
}

func TestFormatMessage(t *testing.T) {
	logger := &Logger{
		level:  InfoLevel,
		format: "text",
	}

	tests := []struct {
		name   string
		format string
		level  Level
		msg    string
		args   []interface{}
	}{
		{
			name:   "text format",
			format: "text",
			level:  InfoLevel,
			msg:    "test message",
			args:   []interface{}{},
		},
		{
			name:   "text format with args",
			format: "text",
			level:  ErrorLevel,
			msg:    "error: %s",
			args:   []interface{}{"something went wrong"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger.format = tt.format
			result := logger.formatMessage(tt.level, tt.msg, tt.args...)

			// Check that the message contains the level
			if !strings.Contains(result, tt.level.String()) {
				t.Errorf("formatMessage() does not contain level %s: %s", tt.level.String(), result)
			}

			// Check that the message contains the expected text
			expectedMsg := tt.msg
			if len(tt.args) > 0 {
				// For messages with arguments, just check the base format
				if !strings.Contains(result, strings.Split(tt.msg, "%")[0]) {
					t.Errorf("formatMessage() does not contain expected message part: %s", result)
				}
			} else {
				if !strings.Contains(result, expectedMsg) {
					t.Errorf("formatMessage() does not contain expected message: %s", result)
				}
			}
		})
	}
}

func TestFormatMessageJSON(t *testing.T) {
	logger := &Logger{
		level:  InfoLevel,
		format: "json",
	}

	result := logger.formatMessage(InfoLevel, "test message")

	// Check that the result contains JSON fields
	if !strings.Contains(result, `"level":"INFO"`) {
		t.Errorf("JSON format does not contain level field: %s", result)
	}
	if !strings.Contains(result, `"message":"test message"`) {
		t.Errorf("JSON format does not contain message field: %s", result)
	}
	if !strings.Contains(result, `"time"`) {
		t.Errorf("JSON format does not contain time field: %s", result)
	}
}

func TestLoggingFunctions(t *testing.T) {
	tests := []struct {
		name      string
		logFunc   func(string, ...interface{})
		level     Level
		message   string
		logLevel  Level
		shouldLog bool
	}{
		{"Debug logs at debug level", Debug, DebugLevel, "debug message", DebugLevel, true},
		{"Debug doesn't log at info level", Debug, InfoLevel, "debug message", DebugLevel, false},
		{"Info logs at info level", Info, InfoLevel, "info message", InfoLevel, true},
		{"Info doesn't log at warn level", Info, WarnLevel, "info message", InfoLevel, false},
		{"Warn logs at warn level", Warn, WarnLevel, "warn message", WarnLevel, true},
		{"Error logs at error level", Error, ErrorLevel, "error message", ErrorLevel, true},
		{"Error logs at debug level", Error, DebugLevel, "error message", ErrorLevel, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to capture output
			var buf bytes.Buffer

			// Set up the logger with the buffer
			defaultLogger = &Logger{
				level:         tt.level,
				format:        "text",
				consoleWriter: &buf,
			}

			// Call the logging function
			tt.logFunc(tt.message)

			// Check if output was written
			output := buf.String()
			if tt.shouldLog {
				if output == "" {
					t.Errorf("Expected log output but got none")
				}
				if !strings.Contains(output, tt.message) {
					t.Errorf("Log output does not contain expected message: %s", output)
				}
			} else if output != "" {
				t.Errorf("Expected no log output but got: %s", output)
			}
		})
	}
}

func TestLoggingWithFormatting(t *testing.T) {
	var buf bytes.Buffer
	defaultLogger = &Logger{
		level:         InfoLevel,
		format:        "text",
		consoleWriter: &buf,
	}

	Info("test %s with %d", "message", 42)

	output := buf.String()
	if !strings.Contains(output, "test message with 42") {
		t.Errorf("Log output does not contain formatted message: %s", output)
	}
}
