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

	"kafka_go_cli/internal/model"
	"kafka_go_cli/internal/processor"
)

// Load reads configuration from multiple sources (config file, environment, CLI flags)
// and returns a Settings struct with all values merged according to precedence:
// defaults < config file < environment variables < CLI flags
func Load(cmd *cobra.Command) (model.Settings, error) {
	// viper will manage the config values
	// for all sources except CLI flags,
	// which we will bind manually
	v := viper.New()
	setDefaultsForGeneralProperties(v)

	// set the config file if provided via CLI flag
	configPath, _ := cmd.Root().PersistentFlags().GetString("config")
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			// When a config path is explicitly provided, failing to read it is an error.
			return model.Settings{}, err
		}
	}

	// set the environment variable prefix
	v.SetEnvPrefix("KAFKA_GO_CLI")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	// IMPORTANT: Only bind flags on CLI that the user explicitly set.
	// If we bind all flags, their default values override config/env values,
	// breaking the intended precedence: defaults < config file < env < CLI.
	bindIfChanged(v, "log_level", lookupFlagOnCLI(cmd, "log-level"))
	bindIfChanged(v, "message_location", lookupFlagOnCLI(cmd, "message-location"))
	bindIfChanged(v, "run_once", lookupFlagOnCLI(cmd, "run-once"))
	bindIfChanged(v, "no_delete_files", lookupFlagOnCLI(cmd, "no-delete-files"))
	bindIfChanged(v, "delay", lookupFlagOnCLI(cmd, "delay"))
	bindIfChanged(v, "max_cycles", lookupFlagOnCLI(cmd, "max-cycles"))
	bindIfChanged(v, "processor", lookupFlagOnCLI(cmd, "processor"))

	var settings model.Settings
	if err := v.Unmarshal(&settings); err != nil {
		return model.Settings{}, err
	}
	settings.ConfigFile = v.ConfigFileUsed()

	// IMPORTANT: at this point we have only handled the general
	// configuration properties that are common to all processors.
	// Now we need to handle processor-specific configuration properties
	// Extract processor-specific configuration from config file
	// e.g., if processor is "kafka", extract the "kafka" key from config
	if settings.ProcessorType != "" {
		processorConfig := v.GetStringMap(settings.ProcessorType)
		if processorConfig != nil {
			settings.ProcessorConfig = processorConfig
		} else {
			settings.ProcessorConfig = make(map[string]interface{})
		}
	}

	// get the processor-specific config params that were registered by the processor
	// so we know which environment variables and CLI flags to check for
	processorParams := processor.GetConfigParams(settings.ProcessorType)
	if processorParams != nil {
		mergeProcessorEnv(v, &settings, processorParams)
		mergeProcessorFlags(cmd, &settings, processorParams)
	}

	return settings, nil
}

// mergeProcessorEnv merges processor-specific env values into settings' ProcessorConfig.
// This ensures env values override config file values and are overridden by CLI flags.
func mergeProcessorEnv(v *viper.Viper, settings *model.Settings, params []model.ConfigParam) {
	if params == nil {
		return
	}

	for _, param := range params {
		if val := v.GetString(param.Flag); val != "" {
			settings.ProcessorConfig[param.Name] = val
		}
	}
}

// mergeProcessorFlags merges processor-specific CLI flags into settings' ProcessorConfig.
// This ensures CLI flags override config file values (implements the proper precedence).
func mergeProcessorFlags(cmd *cobra.Command, settings *model.Settings, params []model.ConfigParam) {
	if params == nil {
		return
	}

	// For each config param, check if the user provided a CLI flag value
	for _, param := range params {
		flag := lookupFlagOnCLI(cmd, param.Flag)
		if flag != nil && flag.Changed {
			// Use the processor-defined name as the config key
			if val := flag.Value.String(); val != "" {
				settings.ProcessorConfig[param.Name] = val
			}
		}
	}
}

func setDefaultsForGeneralProperties(v *viper.Viper) {
	v.SetDefault("config", "")
	v.SetDefault("log_level", "info")
	v.SetDefault("message_location", "")
	v.SetDefault("run_once", false)
	v.SetDefault("no_delete_files", true)
	v.SetDefault("delay", 1000)
	v.SetDefault("max_cycles", -1)
	v.SetDefault("processor", "noop")
}

func bindIfChanged(v *viper.Viper, key string, cliFlagValue *pflag.Flag) {
	if cliFlagValue == nil {
		// flag not found on CLI
		return
	}
	if !cliFlagValue.Changed {
		// flag exists but user did not set it (default), so do not bind
		return
	}
	// bind the flag to viper if it was explicitly set by the user
	// allowing it to override config file & environment variables
	_ = v.BindPFlag(key, cliFlagValue)
}

func lookupFlagOnCLI(cmd *cobra.Command, name string) *pflag.Flag {
	if flag := cmd.Flags().Lookup(name); flag != nil {
		return flag
	}
	return cmd.Root().PersistentFlags().Lookup(name)
}
