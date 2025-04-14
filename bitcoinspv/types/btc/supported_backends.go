package btc

type (
	// SupportedNetwork represents a supported Bitcoin network type (mainnet, testnet, etc.)
	SupportedNetwork string
	// SupportedBackend represents a supported Bitcoin backend implementation (btcd, bitcoind)
	SupportedBackend string
)

// Constants defining the supported Bitcoin networks.
const (
	Mainnet SupportedNetwork = "mainnet"
	Testnet SupportedNetwork = "testnet"
	Simnet  SupportedNetwork = "simnet"
	Regtest SupportedNetwork = "regtest"
	Signet  SupportedNetwork = "signet"
)

// Constants defining the supported Bitcoin backend implementations.
const (
	Btcd     SupportedBackend = "btcd"
	Bitcoind SupportedBackend = "bitcoind"
)

func (n SupportedNetwork) String() string {
	return string(n)
}

func (b SupportedBackend) String() string {
	return string(b)
}

// GetValidNetParams returns a map of valid Bitcoin network parameters
// TODO: Simplify this function !!!
func GetValidNetParams() map[string]bool {
	return map[string]bool{
		Mainnet.String(): true,
		Testnet.String(): true,
		Simnet.String():  true,
		Regtest.String(): true,
		Signet.String():  true,
	}
}

// TODO: Simplify this function !!!
// GetValidBackends returns a map of supported Bitcoin backend types
func GetValidBackends() map[SupportedBackend]bool {
	return map[SupportedBackend]bool{
		Bitcoind: true,
		Btcd:     true,
	}
}
