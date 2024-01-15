package configs

import (
	"sync"

	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// ConfigurationProvider interface to isolation configs
type ConfigurationProvider interface {
	// Get Configuration from ConfigurationProvider
	Get() *Configuration
}

// ConfigurationProviderImpl is implements for ConfigurationProvider
type ConfigurationProviderImpl struct {
	Content *Configuration
	viper   *viper.Viper
	mu      sync.RWMutex
}

// Get is implements for ConfigurationProvider
func (c *ConfigurationProviderImpl) Get() *Configuration {
	return c.Content
}

// NewConfig read configs and create new instance
func NewConfig(path string) (ConfigurationProvider, error) {
	// set file type toml or yaml
	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigType("yaml")

	var config Configuration

	// user doesn't input any configs
	// then check env export configs path
	if path != "" {
		// direct set config path
		v.SetConfigFile(path)
		log.Info().Msgf("server using %s", v.ConfigFileUsed())
	} else {
		name := v.GetString("CONFIG_NAME")
		if name == "" {
			name = "app"
		}
		v.SetConfigName(name)
		v.AddConfigPath("/etc/mome")
		v.AddConfigPath("$HOME/.mome")
		v.AddConfigPath("./deployments/config")
		v.AddConfigPath("./")
		log.Info().Msgf("find config in defalut paths")
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	err := v.Unmarshal(&config, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
		ExpandEnvHook,
	)))
	if err != nil {
		return nil, err
	}

	ci := &ConfigurationProviderImpl{
		Content: &config,
		viper:   v,
	}

	return ci, nil
}
