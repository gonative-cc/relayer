package types

type (
	SupportedBtcNetwork string
	SupportedBtcBackend string
)

const (
	BtcMainnet SupportedBtcNetwork = "mainnet"
	BtcTestnet SupportedBtcNetwork = "testnet"
	BtcSimnet  SupportedBtcNetwork = "simnet"
	BtcRegtest SupportedBtcNetwork = "regtest"
	BtcSignet  SupportedBtcNetwork = "signet"

	Btcd     SupportedBtcBackend = "btcd"
	Bitcoind SupportedBtcBackend = "bitcoind"
)

func (c SupportedBtcNetwork) String() string {
	return string(c)
}

func (c SupportedBtcBackend) String() string {
	return string(c)
}
