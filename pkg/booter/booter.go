package booter

import (
	"flag"
	"os"
	"path"
	"strings"

	"github.com/chan-jui-huang/go-backend-package/v2/pkg/booter/config"
	"github.com/spf13/viper"
)

type Config struct {
	RootDir        string
	ConfigFileName string
	Debug          bool
}

func NewConfig(rootDir string, configFileName string, debug bool) *Config {
	return &Config{
		RootDir:        rootDir,
		ConfigFileName: configFileName,
		Debug:          debug,
	}
}

func NewConfigWithCommand() *Config {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	var rootDir string
	var configFileName string
	var debug bool
	flag.StringVar(&rootDir, "rootDir", wd, "root directory which the executable file in")
	flag.StringVar(&configFileName, "configFileName", "config.yml", "config file name")
	flag.BoolVar(&debug, "debug", false, "debug mode")
	flag.Parse()

	return NewConfig(rootDir, configFileName, debug)
}

func NewDefaultConfig() *Config {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return NewConfig(wd, "config.yml", false)
}

func BootConfigLoader(booterConfig *Config) *config.Loader {
	byteYaml, err := os.ReadFile(path.Join(booterConfig.RootDir, booterConfig.ConfigFileName))
	if err != nil {
		panic(err)
	}
	stringYaml := os.ExpandEnv(string(byteYaml))

	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(strings.NewReader(stringYaml)); err != nil {
		panic(err)
	}

	return config.New(v)
}
