/*
The Config package handles the application configuration. Configurations can come from a variety of places, and
are listed below in order of precedence:
  - Command Line
  - .ecg.yaml
  - .ecg/config.yaml
  - ~/.ecg.yaml
  - <XDG_CONFIG_HOME>/ecg/config.yaml
  - Environment Variables prefixed with ECG_
*/
package config

import (
	"fmt"
	"path"
	"strings"

	"github.com/adrg/xdg"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"

	"github.com/anchore/elastic-container-gatherer/internal"
)

const redacted = "******"

// Configuration options that may only be specified on the command line
type CliOnlyOptions struct {
	ConfigPath string
	Verbosity  int
}

// All Application configurations
type Application struct {
	ConfigPath             string
	Log                    Logging `mapstructure:"log"`
	CliOptions             CliOnlyOptions
	PollingIntervalSeconds int         `mapstructure:"polling-interval-seconds"`
	AnchoreDetails         AnchoreInfo `mapstructure:"anchore"`
	Region                 string      `mapstructure:"region"`
}

// Information for posting in-use image details to Anchore (or any URL for that matter)
type AnchoreInfo struct {
	URL      string     `mapstructure:"url"`
	User     string     `mapstructure:"user"`
	Password string     `mapstructure:"password"`
	Account  string     `mapstructure:"account"`
	HTTP     HTTPConfig `mapstructure:"http"`
}

// Configurations for the HTTP Client itself (net/http)
type HTTPConfig struct {
	Insecure       bool `mapstructure:"insecure"`
	TimeoutSeconds int  `mapstructure:"timeout-seconds"`
}

// Logging Configuration
type Logging struct {
	Level        string `mapstructure:"level"`
	FileLocation string `mapstructure:"file"`
}

// Return whether or not AnchoreDetails are specified
func (anchore *AnchoreInfo) IsValid() bool {
	return anchore.URL != "" &&
		anchore.User != "" &&
		anchore.Password != ""
}

func setNonCliDefaultValues(v *viper.Viper) {
	v.SetDefault("log.level", "")
	v.SetDefault("log.file", "")
	v.SetDefault("anchore.account", "admin")
	v.SetDefault("anchore.http.insecure", false)
	v.SetDefault("anchore.http.timeout-seconds", 10)
}

// Load the Application Configuration from the Viper specifications
func LoadConfigFromFile(v *viper.Viper, cliOpts *CliOnlyOptions) (*Application, error) {
	// the user may not have a config, and this is OK, we can use the default config + default cobra cli values instead
	setNonCliDefaultValues(v)
	if cliOpts != nil {
		_ = readConfig(v, cliOpts.ConfigPath, internal.ApplicationName)
	} else {
		_ = readConfig(v, "", internal.ApplicationName)
	}

	config := &Application{
		CliOptions: *cliOpts,
	}
	err := v.Unmarshal(config)
	if err != nil {
		return nil, fmt.Errorf("unable to parse config: %w", err)
	}
	config.ConfigPath = v.ConfigFileUsed()

	err = config.Build()
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return config, nil
}

// Build the configuration object (to be used as a singleton)
func (cfg *Application) Build() error {
	if cfg.Log.Level != "" {
		if cfg.CliOptions.Verbosity > 0 {
			return fmt.Errorf("cannot explicitly set log level (cfg file or env var) and use -v flag together")
		}
	} else {
		switch v := cfg.CliOptions.Verbosity; {
		case v == 1:
			cfg.Log.Level = "info"
		case v >= 2:
			cfg.Log.Level = "debug"
		default:
			cfg.Log.Level = "error"
		}
	}

	return nil
}

func readConfig(v *viper.Viper, configPath, applicationName string) error {
	v.AutomaticEnv()
	v.SetEnvPrefix(applicationName)
	// allow for nested options to be specified via environment variables
	// e.g. pod.context = APPNAME_POD_CONTEXT
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// use explicitly the given user config
	if configPath != "" {
		fmt.Println("using config file:", configPath)
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err == nil {
			return nil
		}
		// don't fall through to other options if this fails
		return fmt.Errorf("unable to read config: %v", configPath)
	}

	// start searching for valid configs in order...

	// 1. look for .<appname>.yaml (in the current directory)
	v.AddConfigPath(".")
	v.SetConfigName(applicationName)
	if err := v.ReadInConfig(); err == nil {
		return nil
	}

	// 2. look for .<appname>/config.yaml (in the current directory)
	v.AddConfigPath("." + applicationName)
	v.SetConfigName("config")
	if err := v.ReadInConfig(); err == nil {
		return nil
	}

	// 3. look for ~/.<appname>.yaml
	home, err := homedir.Dir()
	if err == nil {
		v.AddConfigPath(home)
		v.SetConfigName("." + applicationName)
		if err := v.ReadInConfig(); err == nil {
			return nil
		}
	}

	// 4. look for <appname>/config.yaml in xdg locations (starting with xdg home config dir, then moving upwards)
	v.AddConfigPath(path.Join(xdg.ConfigHome, applicationName))
	for _, dir := range xdg.ConfigDirs {
		v.AddConfigPath(path.Join(dir, applicationName))
	}
	v.SetConfigName("config")
	if err := v.ReadInConfig(); err == nil {
		return nil
	}

	return fmt.Errorf("application config not found")
}

func (cfg Application) String() string {
	// redact sensitive information
	// Note: If the configuration grows to have more redacted fields it would be good to refactor this into something that
	// is more dynamic based on a property or list of "sensitive" fields
	if cfg.AnchoreDetails.Password != "" {
		cfg.AnchoreDetails.Password = redacted
	}

	// yaml is pretty human friendly (at least when compared to json)
	appCfgStr, err := yaml.Marshal(&cfg)

	if err != nil {
		return err.Error()
	}

	return string(appCfgStr)
}
