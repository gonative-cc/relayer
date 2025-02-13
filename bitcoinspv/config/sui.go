package config

type SuiConfig struct {
	Endpoint   string `mapstructure:"endpoint"`
	Mnemonic   string `mapstructure:"mnemonic"`
	LCObjectId string `mapstructure:"lc_object_id"`
}
