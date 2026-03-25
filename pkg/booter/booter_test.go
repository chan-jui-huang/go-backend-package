package booter_test

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	booter "github.com/chan-jui-huang/go-backend-package/v2/pkg/booter"
)

func TestNewConfig(t *testing.T) {
	cfg := booter.NewConfig("/tmp/project", "custom.yml", true)

	if cfg.RootDir != "/tmp/project" {
		t.Fatalf("expected RootDir to be /tmp/project, got %s", cfg.RootDir)
	}
	if cfg.ConfigFileName != "custom.yml" {
		t.Fatalf("expected ConfigFileName to be custom.yml, got %s", cfg.ConfigFileName)
	}
	if !cfg.Debug {
		t.Fatalf("expected Debug to be true")
	}
}

func TestNewDefaultConfig(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	cfg := booter.NewDefaultConfig()

	if cfg.RootDir != wd {
		t.Fatalf("expected RootDir to be %s, got %s", wd, cfg.RootDir)
	}
	if cfg.ConfigFileName != "config.yml" {
		t.Fatalf("expected ConfigFileName to be config.yml, got %s", cfg.ConfigFileName)
	}
	if cfg.Debug {
		t.Fatalf("expected Debug to be false")
	}
}

func TestNewConfigWithCommand(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	originalArgs := os.Args
	originalCommandLine := flag.CommandLine
	defer func() {
		os.Args = originalArgs
		flag.CommandLine = originalCommandLine
	}()

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = []string{
		"cmd",
		"-rootDir=/tmp/cli-project",
		"-configFileName=cli.yml",
		"-debug=true",
	}

	cfg := booter.NewConfigWithCommand()

	if cfg.RootDir != "/tmp/cli-project" {
		t.Fatalf("expected RootDir to be /tmp/cli-project, got %s", cfg.RootDir)
	}
	if cfg.ConfigFileName != "cli.yml" {
		t.Fatalf("expected ConfigFileName to be cli.yml, got %s", cfg.ConfigFileName)
	}
	if !cfg.Debug {
		t.Fatalf("expected Debug to be true")
	}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = []string{"cmd"}

	defaultCfg := booter.NewConfigWithCommand()

	if defaultCfg.RootDir != wd {
		t.Fatalf("expected default RootDir to be %s, got %s", wd, defaultCfg.RootDir)
	}
	if defaultCfg.ConfigFileName != "config.yml" {
		t.Fatalf("expected default ConfigFileName to be config.yml, got %s", defaultCfg.ConfigFileName)
	}
	if defaultCfg.Debug {
		t.Fatalf("expected default Debug to be false")
	}
}

func TestBootConfigLoaderEnvExpansion(t *testing.T) {
	tmp := t.TempDir()
	cfg := "database:\n  host: \"${MY_HOST}\"\n"
	if err := os.WriteFile(filepath.Join(tmp, "config.yml"), []byte(cfg), 0644); err != nil {
		t.Fatal(err)
	}
	// set env
	os.Setenv("MY_HOST", "localhost")
	defer os.Unsetenv("MY_HOST")

	bc := booter.NewConfig(tmp, "config.yml", false)
	loader := booter.BootConfigLoader(bc)

	type databaseConfig struct {
		Host string `mapstructure:"host"`
	}

	cfgObj := &databaseConfig{}
	loader.Unmarshal("database", cfgObj)
	if cfgObj.Host != "localhost" {
		t.Fatalf("expected database.host to be 'localhost', got '%s'", cfgObj.Host)
	}
}

func TestBootConfigLoaderPanicWhenFileNotFound(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic when config file does not exist")
		}
	}()

	booter.BootConfigLoader(booter.NewConfig(t.TempDir(), "missing.yml", false))
}

func TestBootConfigLoaderPanicWhenYamlInvalid(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "config.yml"), []byte("database: ["), 0644); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic when yaml is invalid")
		}
	}()

	booter.BootConfigLoader(booter.NewConfig(tmp, "config.yml", false))
}
