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

import "io"

// Config define setup log output setting
type Config struct {
	// Env is for log output tag
	// each log will auto add this tag when this field is not empty
	Env string `mapstructure:"env"`

	// App is for log output tag
	// each log will auto add this tag when this field is not empty
	App string `mapstructure:"app"`

	// Level define this logger level
	// accept level, see level definition
	Level Level `mapstructure:"level"`

	// Debug is control pretty log output
	// the flag is useful local debug
	Debug bool `mapstructure:"debug"`

	// logs Directory where grafana can store logs
	Path string `mapstructure:"path"`

	// writer extends log output
	writer io.Writer
}

// An Option is passed to Config
type Option interface {
	apply(*Config)
}

// setEnv for implement config option pattern
type setEnv struct{ env string }

// apply implement Option interface
func (opt *setEnv) apply(c *Config) { c.Env = opt.env }

// setApp for implement config option pattern
type setApp struct{ app string }

// apply implement Option interface
func (opt *setApp) apply(c *Config) { c.App = opt.app }

// setDebug for implement config option pattern
type setDebug struct{ debug bool }

// apply implement Option interface
func (opt *setDebug) apply(c *Config) { c.Debug = opt.debug }

// setLevel for implement config option pattern
type setLevel struct{ level Level }

// apply implement Option interface
func (opt *setLevel) apply(c *Config) { c.Level = opt.level }

// setWriter for implement config option pattern
type setWriter struct{ writer io.Writer }

// apply implement Option interface
func (opt *setWriter) apply(c *Config) { c.writer = opt.writer }

// WithEnv with env log tag
func WithEnv(env string) Option {
	return &setEnv{
		env: env,
	}
}

// WithApp with app log tag
func WithApp(app string) Option {
	return &setApp{
		app: app,
	}
}

// WithDebug with debug flag
func WithDebug(debug bool) Option {
	return &setDebug{
		debug: debug,
	}
}

// WithLevel with log level
func WithLevel(level Level) Option {
	return &setLevel{
		level: level,
	}
}

// WithOutput with output option
func WithOutput(w io.Writer) Option {
	return &setWriter{writer: w}
}
