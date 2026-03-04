package logger

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap/zapcore"
)

func TestNewConsoleLoggerEncoders(t *testing.T) {
	tests := []struct {
		name       string
		encoder    zapcore.Encoder
		wantSubstr string
	}{
		{"json", JsonEncoder, "\"message\""},
		{"console", ConsoleEncoder, "hello-console"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// capture stdout to temp file
			tmpfile, err := os.CreateTemp("", "stdoutlog")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())
			orig := os.Stdout
			os.Stdout = tmpfile
			defer func() { os.Stdout = orig }()

			cfg := Config{Type: Console, Level: Info}
			logger, err := NewLogger(cfg, tt.encoder)
			if err != nil {
				t.Fatal(err)
			}
			defer logger.Sync()

			// write a message
			if tt.name == "json" {
				logger.Info("hello-json")
			} else {
				logger.Info("hello-console")
			}

			// ensure writes flushed
			_ = tmpfile.Sync()
			_ = tmpfile.Close()

			b, err := ioutil.ReadFile(tmpfile.Name())
			if err != nil {
				t.Fatal(err)
			}
			s := string(b)
			if !strings.Contains(s, tt.wantSubstr) {
				t.Fatalf("expected stdout to contain %q; got: %s", tt.wantSubstr, s)
			}
		})
	}
}

func TestNewFileLoggerWritesFile(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "logdir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	logPath := filepath.Join(tmpdir, "test.log")
	cfg := Config{
		Type:       File,
		LogPath:    logPath,
		MaxSize:    1024,
		MaxBackups: 1,
		MaxAge:     time.Hour,
		Compress:   false,
		Level:      Info,
	}
	logger, err := NewLogger(cfg, JsonEncoder)
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Sync()

	logger.Info("file-log-message")
	// give some time for write
	time.Sleep(10 * time.Millisecond)

	b, err := ioutil.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "file-log-message") {
		t.Fatalf("expected log file to contain message; got: %s", string(b))
	}
}

func TestSamplingApplied(t *testing.T) {
	cfg := Config{Type: Console, Level: Info}
	// no sampler
	loggerNoSampler := NewConsoleLogger(cfg, JsonEncoder)
	defer loggerNoSampler.Sync()
	coreNo := reflect.TypeOf(loggerNoSampler.Core()).String()

	// with sampler
	loggerWithSampler, err := NewLogger(cfg, JsonEncoder, DefaultZapOptions...)
	if err != nil {
		t.Fatal(err)
	}
	defer loggerWithSampler.Sync()
	coreWith := reflect.TypeOf(loggerWithSampler.Core()).String()

	if strings.Contains(coreNo, "sampler") {
		t.Fatalf("expected core without options to not be a sampler, got: %s", coreNo)
	}
	if !strings.Contains(coreWith, "sampler") {
		t.Fatalf("expected core with DefaultZapOptions to be a sampler, got: %s", coreWith)
	}
}
