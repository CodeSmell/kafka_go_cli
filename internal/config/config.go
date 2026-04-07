/*
The configuration for the application is loaded using the `Load` function in the `config` package.
This function uses Viper to read configuration values from multiple sources, including:
- Command-line flags
- Environment variables
- Configuration files
*/
package config

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Settings struct {
	LogLevel        string `mapstructure:"log_level" json:"log_level"`
	MessageLocation string `mapstructure:"message_location" json:"message_location"`
	RunOnce         bool   `mapstructure:"run_once" json:"run_once"`
	NoDeleteFiles   bool   `mapstructure:"no_delete_files" json:"no_delete_files"`
	ConfigFile      string `json:"config_file"`
}

func Load(cmd *cobra.Command) (Settings, error) {
	v := viper.New()
	setDefaults(v)

	v.SetEnvPrefix("KAFKA_GO_CLI")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	// IMPORTANT: Only bind flags that the user explicitly set.
	// If we bind all flags, their default values override config/env values,
	// breaking the intended precedence: defaults < config file < env < CLI.
	bindIfChanged(v, "log_level", lookupFlag(cmd, "log-level"))
	bindIfChanged(v, "message_location", lookupFlag(cmd, "message-location"))
	bindIfChanged(v, "run_once", lookupFlag(cmd, "run-once"))
	bindIfChanged(v, "no_delete_files", lookupFlag(cmd, "no-delete-files"))

	configPath, _ := cmd.Root().PersistentFlags().GetString("config")
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			// When a config path is explicitly provided, failing to read it is an error.
			return Settings{}, err
		}
	}

	var settings Settings
	if err := v.Unmarshal(&settings); err != nil {
		return Settings{}, err
	}
	settings.ConfigFile = v.ConfigFileUsed()

	return settings, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("log_level", "info")
	v.SetDefault("message_location", "")
	v.SetDefault("run_once", false)
	v.SetDefault("no_delete_files", true)
}

func bindIfChanged(v *viper.Viper, key string, flag *pflag.Flag) {
	if flag == nil {
		return
	}
	if !flag.Changed {
		return
	}
	_ = v.BindPFlag(key, flag)
}

func lookupFlag(cmd *cobra.Command, name string) *pflag.Flag {
	if flag := cmd.Flags().Lookup(name); flag != nil {
		return flag
	}
	return cmd.Root().PersistentFlags().Lookup(name)
}
