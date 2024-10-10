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

package config

import (
	"embed"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"gsmate/internal/utils"

	syslocale "github.com/jeandeaual/go-locale"
	"github.com/pkg/errors"
	"github.com/vimiix/pkg/file"
	"github.com/xo/terminfo"
	"gopkg.in/ini.v1"
)

//go:embed defaultconfig.ini
var defaultConfigFile embed.FS

var (
	defaultConfig *Config
	printConfig   map[string]string
)

const defaultPrompt = "$u@$h/$d> "

func Get() *Config {
	return defaultConfig
}

func GetPrintConfig() map[string]string {
	return printConfig
}

type Config struct {
	Prompt               string `ini:"prompt,omitempty"`
	LessChatty           bool   `ini:"less_chatty,omitempty"`
	MaxHistory           int    `ini:"max_history,omitempty"`
	LogLevel             string `ini:"log_level,omitempty"`
	Silence              bool   `ini:"silence,omitempty"`
	SyntaxHighlight      bool   `ini:"syntax_highlight,omitempty"`
	SyntaxHighlightStyle string `ini:"syntax_highlight_style,omitempty"`
	OnErrorStop          bool   `ini:"on_error_stop,omitempty"`

	// auto detected fields
	Pager                 string `ini:"-"`
	Editor                string `ini:"-"`
	SyntaxHighlightFormat string `ini:"-"`
	SSLMode               string `ini:"-"`

	Connection `ini:"connection"`
}

func GetConfigMap() map[string]string {
	c := defaultConfig
	return map[string]string{
		"prompt":                 c.Prompt,
		"less_chatty":            strconv.FormatBool(c.LessChatty),
		"max_history":            strconv.Itoa(c.MaxHistory),
		"log_level":              c.LogLevel,
		"silence":                strconv.FormatBool(c.Silence),
		"syntax_highlight":       strconv.FormatBool(c.SyntaxHighlight),
		"syntax_highlight_style": c.SyntaxHighlightStyle,
		"on_error_stop":          strconv.FormatBool(c.OnErrorStop),
	}
}

func (c *Config) LivePrompt() func() (string, bool) {
	return func() (string, bool) {
		if c.Prompt == "" {
			c.Prompt = defaultPrompt
		}

		rs := []rune(c.Prompt)
		var buf []byte
		end := len(rs)
		for i := 0; i < len(rs); i++ {
			if rs[i] != '$' {
				buf = append(buf, string(rs[i])...)
				continue
			}

			switch utils.Grab(rs, i+1, end) {
			case '$':
				buf = append(buf, '$')
			case 'u':
				buf = append(buf, []byte(c.Connection.Username)...)
			case 'h':
				buf = append(buf, []byte(c.Connection.Host)...)
			case 'd':
				buf = append(buf, []byte(c.Connection.DBName)...)
			case 'p':
				buf = append(buf, []byte(strconv.Itoa(c.Connection.Port))...)
			case 'i':
				pid := os.Getpid()
				buf = append(buf, []byte(strconv.Itoa(pid))...)
				// TODO support more
			default:
			}
			i++
		}
		return string(buf), true
	}
}

func Init() error {
	defaultConfig = newDefault()
	cfgFile := filepath.Join(DefaultLocation(), "config")
	if err := writeDefaultConfig(cfgFile, false); err != nil {
		return err
	}
	if err := ini.MapTo(defaultConfig, cfgFile); err != nil {
		return errors.Wrapf(err, "load config: %s", cfgFile)
	}

	locale := "en-US"
	if s, err := syslocale.GetLocale(); err == nil {
		locale = s
	}
	pager := "off"
	if defaultConfig.Pager != "" {
		pager = "on"
	}
	printConfig = map[string]string{
		"border":                   "1",
		"columns":                  "0",
		"csv_fieldsep":             ",",
		"expanded":                 "off",
		"fieldsep":                 "|",
		"fieldsep_zero":            "off",
		"footer":                   "on",
		"format":                   "aligned",
		"linestyle":                "ascii",
		"locale":                   locale,
		"null":                     "",
		"numericlocale":            "off",
		"pager_min_lines":          "0",
		"pager":                    pager,
		"pager_cmd":                defaultConfig.Pager,
		"recordsep":                "\n",
		"recordsep_zero":           "off",
		"tableattr":                "",
		"time":                     "RFC3339Nano",
		"timezone":                 "",
		"title":                    "",
		"tuples_only":              "off",
		"unicode_border_linestyle": "single",
		"unicode_column_linestyle": "single",
		"unicode_header_linestyle": "single",
	}
	return nil
}

func newDefault() *Config {
	noColor := false
	if s, ok := utils.Getenv("NO_COLOR"); ok {
		noColor = s != "0" && s != "false" && s != "off"
	}
	colorLevel, _ := terminfo.ColorLevelFromEnv()
	enableHighlight := true
	if noColor || colorLevel < terminfo.ColorLevelBasic {
		enableHighlight = false
	}
	pagerCmd, ok := utils.Getenv("PAGER")
	if !ok {
		for _, s := range []string{"less", "more"} {
			if _, err := exec.LookPath(s); err == nil {
				pagerCmd = s
				break
			}
		}
	}

	editorCmd, _ := utils.Getenv("EDITOR", "VISUAL")
	sslmode, ok := utils.Getenv("SSLMODE")
	if !ok {
		sslmode = "retry"
	}
	return &Config{
		Prompt:                defaultPrompt,
		MaxHistory:            1000,
		LogLevel:              "info",
		SyntaxHighlight:       enableHighlight,
		SyntaxHighlightStyle:  "monokai",
		SyntaxHighlightFormat: colorLevel.ChromaFormatterName(),

		Pager:   pagerCmd,
		Editor:  editorCmd,
		SSLMode: sslmode,
		Connection: Connection{
			Host:         "localhost",
			Port:         26000,
			Username:     "omm",
			DBName:       "postgres",
			ConnTimeout:  time.Second * 10,
			QueryTimeout: time.Second * 120,
		},
	}
}

// DefaultLocation returns the default location of the config file, which is
// determined by the XDG configuration directory specification. On Windows,
// the configuration directory is in the user's AppData directory. If the
// XDG_CONFIG_HOME environment variable is not set, the default location is
// ~/.config/gsmate/ on Unix systems and %USERPROFILE%\AppData\Local\gsmate\
// on Windows.
func DefaultLocation() string {
	if os.Getenv("XDG_CONFIG_HOME") != "" {
		return file.ExpandHomePath(os.Getenv("XDG_CONFIG_HOME")) + "/gsmate/"
	}
	if runtime.GOOS == "windows" {
		return os.Getenv("USERPROFILE") + "\\AppData\\Local\\gsmate\\"
	}
	return file.ExpandHomePath("~/.config/gsmate/")
}

func writeDefaultConfig(dest string, overwrite bool) error {
	dest = file.ExpandHomePath(dest)
	if !overwrite && file.Exists(dest) {
		return nil
	}

	if err := file.EnsureDirExists(dest); err != nil {
		return err
	}

	src, err := defaultConfigFile.Open("defaultconfig.ini")
	if err != nil {
		return err
	}
	dst, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	return err
}
