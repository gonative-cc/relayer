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
)

const (
	Btcd     SupportedBtcBackend = "btcd"
	Bitcoind SupportedBtcBackend = "bitcoind"
)

func (n SupportedBtcNetwork) String() string {
	return string(n)
}

func (b SupportedBtcBackend) String() string {
	return string(b)
}
