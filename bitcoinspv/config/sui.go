package config

type SuiConfig struct {
	Endpoint    string `mapstructure:"endpoint"`
	Mnemonic    string `mapstructure:"mnemonic"`
	LCObjectID  string `mapstructure:"lc_object_id"`
	LCPackageID string `mapstructure:"lc_package_id"`
}
