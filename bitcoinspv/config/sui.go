package config

// SuiConfig holds configuration for interacting with the light client on Sui.
type SuiConfig struct {
	Endpoint    string `mapstructure:"endpoint"`
	Mnemonic    string `mapstructure:"mnemonic"`
	LCObjectID  string `mapstructure:"lc_object_id"`
	LCPkgID     string `mapstructure:"lc_package_id"`
	BTCLibPkgID string `mapstructure:"btc_lib_pkg_id"`
}
