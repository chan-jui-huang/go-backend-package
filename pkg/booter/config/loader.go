package config

import (
	"github.com/spf13/viper"
)

type Loader struct {
	viper *viper.Viper
}

func New(v *viper.Viper) *Loader {
	return &Loader{
		viper: v,
	}
}

func (r *Loader) Unmarshal(key string, config any) {
	err := r.viper.UnmarshalKey(key, config)
	if err != nil {
		panic(err)
	}
}

func (r *Loader) UnmarshalMany(configs map[string]any) {
	for key, config := range configs {
		r.Unmarshal(key, config)
	}
}
