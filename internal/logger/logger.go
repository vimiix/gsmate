// Copyright 2024 Qian Yao
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

var (
	slience bool
	level   LogLevel
	logger  = log.New(os.Stderr, "", 0)
)

type LogLevel uint8

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

func SetLogLevelByString(s string) {
	switch strings.ToUpper(s) {
	case "DEBUG":
		level = DebugLevel
	case "INFO":
		level = InfoLevel
	case "WARN":
		level = WarnLevel
	case "ERROR":
		level = ErrorLevel
	case "FATAL":
		level = FatalLevel
	default:
		level = InfoLevel
	}
}

// SetLogLevel sets the log level.
func SetLogLevel(v LogLevel) {
	level = v
}

func MuteLogger() {
	slience = true
}

func defaultPrint(lvl LogLevel, message string) {
	if slience {
		return
	}
	if lvl >= level {
		ts := time.Now().Format("2006-01-02T15:04:05.000")
		logger.Print(
			strings.Join([]string{ts, "[" + lvl.String() + "]", message}, " "),
		)
	}
}

var printFunc = defaultPrint

// convenience functions
var (
	Debugf = Debug
	Infof  = Info
	Warnf  = Warn
	Errorf = Error
	Fatalf = Fatal
)

func Debug(format string, v ...any) {
	printFunc(DebugLevel, fmt.Sprintf(format, v...))
}

func Info(format string, v ...any) {
	printFunc(InfoLevel, fmt.Sprintf(format, v...))
}

func Warn(format string, v ...any) {
	printFunc(WarnLevel, color.YellowString(format, v...))
}

func Error(format string, v ...any) {
	printFunc(ErrorLevel, color.RedString(format, v...))
}

func Fatal(format string, v ...any) {
	printFunc(ErrorLevel, color.RedString(format, v...))
	os.Exit(1)
}
