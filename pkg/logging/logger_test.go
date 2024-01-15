package logging_test

import (
	"bytes"
	"io"
	syslog "log/syslog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/karta0898098/mome/pkg/logging"
)

func Test_LoggerMessageOutput(t *testing.T) {
	var tests = []struct {
		name string
		msg  string
	}{
		{
			name: "Success",
			msg:  "test",
		},
	}
	for _, tt := range tests {
		b := new(bytes.Buffer)
		logger := logging.SetupWithOption(
			logging.WithOutput(io.MultiWriter(b, os.Stdout)),
			logging.WithDebug(true),
		)
		logger.Info().Msg(tt.msg)
		assert.Contains(t, b.String(), tt.msg)
	}
}

func Test_LoggerHasEnvOutput(t *testing.T) {
	var tests = []struct {
		name  string
		env   string
		setup func()
		tear  func()
	}{
		{
			name:  "Success",
			env:   "local",
			setup: func() {},
			tear:  func() {},
		},
		{
			name: "Success",
			env:  "dev",
			setup: func() {
				_ = os.Setenv("LOG_ENV", "dev")
			},
			tear: func() {
				_ = os.Unsetenv("LOG_ENV")
			},
		},
	}

	for _, tt := range tests {
		tt.setup()
		b := new(bytes.Buffer)
		logger := logging.SetupWithOption(
			logging.WithOutput(io.MultiWriter(b, os.Stdout)),
			logging.WithEnv(tt.env),
			logging.WithDebug(true),
		)
		logger.Info().Msg("")
		assert.Contains(t, b.String(), tt.env)
		tt.tear()
	}
}

func Test_LoggerLevel(t *testing.T) {
	var tests = []struct {
		name   string
		msg    string
		level  logging.Level
		hasLog bool
	}{
		{
			name:   "HasLog",
			msg:    "msg",
			level:  logging.DebugLevel,
			hasLog: true,
		},
		{
			name:   "NoLog",
			msg:    "msg",
			level:  logging.ErrorLevel,
			hasLog: false,
		},
	}

	for _, tt := range tests {
		b := new(bytes.Buffer)
		logger := logging.SetupWithOption(
			logging.WithOutput(io.MultiWriter(b, os.Stdout)),
			logging.WithLevel(tt.level),
			logging.WithDebug(true),
		)
		logger.Info().Msg(tt.msg)
		if tt.hasLog {
			assert.Contains(t, b.String(), tt.msg)
		} else {
			assert.Equal(t, b.Len(), 0)
		}
	}
}

func Test_LoggerSyslog(t *testing.T) {
	if os.Getenv("CI_JOB_ID") != "" {
		t.Skip("Skipping testing in CI environment")
	}

	var tests = []struct {
		name string
		msg  string
	}{
		{
			name: "Success",
			msg:  "test msg",
		},
	}

	for _, tt := range tests {
		wr, err := syslog.New(syslog.LOG_DEBUG|syslog.LOG_EMERG, "test")
		assert.NoError(t, err)

		b := new(bytes.Buffer)
		logger := logging.SetupWithOption(
			logging.WithOutput(io.MultiWriter(b, os.Stdout, wr)),
			logging.WithLevel(logging.InfoLevel),
			logging.WithDebug(true),
		)
		logger.Info().Msg(tt.msg)
		assert.Contains(t, b.String(), tt.msg)
	}
}
