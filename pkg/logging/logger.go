/*
 * Copyright (c) 2023.
 * D-Link Corporation.
 * All rights reserved.
 *
 * The information contained herein is confidential and proprietary to
 * D-Link. Use of this information by anyone other than authorized employees
 * of D-Link is granted only under a written non-disclosure agreement,
 * expressly prescribing the scope and manner of such use.
 */

package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	DefaultLoggerConfig = &Config{
		Debug:  false,
		Level:  InfoLevel,
		writer: os.Stdout,
	}
)

// timestampHook handle timestamp format for EFK stack
type timestampHook struct{}

// Run implement zerolog hook
func (h timestampHook) Run(
	e *zerolog.Event,
	level zerolog.Level,
	msg string,
) {
	e.Float64("@timestamp", float64(time.Now().UnixNano()/int64(time.Millisecond))/1000)
}

// SetupWithOption setup logger with option
func SetupWithOption(opts ...Option) zerolog.Logger {
	logger := setup(DefaultLoggerConfig, opts...)
	log.Logger = logger

	return logger
}

// Setup logger
func Setup(cfg Config) zerolog.Logger {
	logger := setup(&cfg)
	log.Logger = logger
	return logger
}

// setup logger function
func setup(config *Config, opts ...Option) zerolog.Logger {
	var (
		logger zerolog.Logger
	)

	if config == nil {
		config = DefaultLoggerConfig
	}

	if config.writer == nil {
		if config.Path != "" {
			writer := DefaultLoggerConfig.writer
			if _, err := os.Stat(config.Path); os.IsNotExist(err) {
				if err := os.MkdirAll(config.Path, 755); err != nil {
					log.Error().
						Err(err).
						Msgf("failed to create folder %v, then using stdout", config.Path)
				}
			}

			logFile, err := os.OpenFile(
				filepath.Join(config.Path, "app.log"),
				os.O_APPEND|os.O_CREATE|os.O_WRONLY,
				0664,
			)
			if err != nil {
				log.Error().Err(err).Msgf("failed to open file, then using stdout")
			} else {
				writer = logFile
			}
			config.writer = writer
		} else {
			config.writer = DefaultLoggerConfig.writer
		}
	}

	// apply options to config
	for _, opt := range opts {
		opt.apply(config)
	}

	zerolog.DisableSampling(true)
	zerolog.TimestampFieldName = "timestamp"
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	level := config.Level

	// setup debug mode logger
	if config.Debug {
		output := zerolog.ConsoleWriter{
			Out: config.writer,
		}
		output.FormatFieldName = func(i interface{}) string {
			return fmt.Sprintf("%s=", Teal(i))
		}
		output.FormatFieldValue = func(i interface{}) string {
			return fmt.Sprintf("%s", i)
		}
		output.FormatTimestamp = func(i interface{}) string {
			t := fmt.Sprintf("%v", i)
			millisecond, err := strconv.ParseInt(fmt.Sprintf("%s", i), 10, 64)
			if err == nil {
				t = time.UnixMilli(millisecond).Format("2006/01/02 15:04:05.000")
			}
			return colorize(t, colorCyan)
		}
		output.FormatCaller = func(i interface{}) string {
			var c string
			if cc, ok := i.(string); ok {
				c = cc
			}
			if len(c) > 0 {
				cwd, err := os.Getwd()
				if err == nil {
					c = strings.TrimPrefix(c, cwd)
					c = strings.TrimPrefix(c, "/")
				}
				c = colorize(c, colorGreen)

				if c != "" {
					c = fmt.Sprintf("%s %s", " >", c)
				}
			}
			return c
		}
		output.PartsOrder = []string{
			zerolog.TimestampFieldName,
			zerolog.LevelFieldName,
			zerolog.MessageFieldName,
			zerolog.CallerFieldName,
		}
		logger = zerolog.New(output)
	} else {
		logger = zerolog.New(config.writer)
	}

	// add app tag to logger
	if config.App != "" {
		logger = logger.With().Str("app", config.App).Logger()
	}

	// add env tag to logger
	if config.Env != "" {
		logger = logger.With().Str("env", config.Env).Logger()
	}

	logger = logger.
		Hook(timestampHook{}).
		With().
		Timestamp().
		Logger().
		Level(zerolog.Level(level))

	log.Logger = logger

	return logger
}

// colorize returns the string s wrapped in ANSI code c, unless disabled is true.
func colorize(s interface{}, c int) string {
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", c, s)
}
