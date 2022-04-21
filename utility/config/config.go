package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config = viper.Viper

func Load(configFile string, configFiles ...string) *Config {
	v := viper.New()
	v.SetConfigName(configFile)
	v.SetConfigType("json")
	v.AddConfigPath(".")
	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("Error config load [%v]: %v\n", configFile, err))
	}
	for _, cf := range configFiles {
		v.SetConfigName(cf)
		if err := v.MergeInConfig(); err != nil {
			panic(fmt.Sprintf("Error config load [%v]: %v\n", cf, err))
		}
	}
	return v
}
