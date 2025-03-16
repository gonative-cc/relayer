package config

// SuiConfig provides parameters for signer and Bitcoin SPV package on Sui to call.
type SuiConfig struct {
	Endpoint    string `mapstructure:"endpoint"`
	Mnemonic    string `mapstructure:"mnemonic"`
	LCObjectID  string `mapstructure:"lc_object_id"`
	LCPackageID string `mapstructure:"lc_package_id"`
}
