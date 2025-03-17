package config

// SuiConfig holds configuration for interacting with the light client on Sui.
type SuiConfig struct {
	Endpoint    string `mapstructure:"endpoint"`
	Mnemonic    string `mapstructure:"mnemonic"`
	LCObjectID  string `mapstructure:"lc_object_id"`
	LCPackageID string `mapstructure:"lc_package_id"`
}
