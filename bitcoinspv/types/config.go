package types

// GetValidNetParams returns a map of valid Bitcoin network parameters
func GetValidNetParams() map[string]bool {
	return map[string]bool{
		BtcMainnet.String(): true,
		BtcTestnet.String(): true,
		BtcSimnet.String():  true,
		BtcRegtest.String(): true,
		BtcSignet.String():  true,
	}
}

// GetValidBtcBackends returns a map of supported Bitcoin backend types
func GetValidBtcBackends() map[SupportedBtcBackend]bool {
	return map[SupportedBtcBackend]bool{
		Bitcoind: true,
		Btcd:     true,
	}
}
