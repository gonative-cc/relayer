module github.com/gonative-cc/relayer

go 1.23.4

toolchain go1.24.1

require (
	github.com/avast/retry-go/v4 v4.6.1
	github.com/block-vision/sui-go-sdk v1.0.7
	github.com/btcsuite/btcd v0.24.2
	github.com/btcsuite/btcd/btcutil v1.1.6
	github.com/fardream/go-bcs v0.8.7
	github.com/joho/godotenv v1.5.1
	github.com/pattonkan/sui-go v0.1.1
	github.com/rs/zerolog v1.34.0
	github.com/spf13/cobra v1.9.1
	github.com/tinylib/msgp v1.3.0
	github.com/vektra/mockery/v2 v2.53.4
)

require (
	github.com/btcsuite/btcutil v1.0.2 // indirect
	github.com/mitchellh/hashstructure/v2 v2.0.2 // indirect
)

require (
	github.com/btcsuite/btcd/btcec/v2 v2.3.4 // indirect
	github.com/btcsuite/btcd/chaincfg/chainhash v1.1.0
	github.com/btcsuite/btclog v0.0.0-20170628155309-84c8d2346e9f // indirect
	github.com/btcsuite/go-socks v0.0.0-20170105172521-4720035b7bfd // indirect
	github.com/btcsuite/websocket v0.0.0-20150119174127-31079b680792 // indirect
	github.com/chigopher/pathlib v0.19.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/decred/dcrd/crypto/blake256 v1.0.1 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.3.0 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.12.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/golang/snappy v0.0.5-0.20220116011046-fa5810519dcb // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/huandu/xstrings v1.4.0 // indirect
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jinzhu/copier v0.4.0 // indirect
	github.com/leodido/go-urn v1.2.2 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20
	github.com/mattn/go-sqlite3 v1.14.28
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/pebbe/zmq4 v1.4.0
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/philhofer/fwd v1.1.3-0.20240916144458-20a13a1f6b7c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	github.com/sagikazarmark/locafero v0.7.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.12.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/spf13/viper v1.20.1
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/stretchr/testify v1.10.0
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20220721030215-126854af5e6d // indirect
	github.com/tidwall/gjson v1.14.4 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tyler-smith/go-bip39 v1.1.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.33.0 // indirect
	golang.org/x/mod v0.23.0 // indirect
	golang.org/x/sync v0.11.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/term v0.29.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	golang.org/x/time v0.8.0 // indirect
	golang.org/x/tools v0.30.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gotest.tools v2.2.0+incompatible
	gotest.tools/v3 v3.5.2
)

replace (
	// temp cosmos-sdk
	github.com/cosmos/cosmos-sdk => github.com/cosmos/cosmos-sdk v0.52.0-rc.1.0.20250106202622-c6327ea6e403
	// use custom version until this PR is merged
	// - https://github.com/strangelove-ventures/cometbft-client/pull/10
	github.com/strangelove-ventures/cometbft-client => github.com/initia-labs/cometbft-client v0.0.0-20240924071428-ef115cefa07e
	// nohooyr repo moved to github/coder
	nhooyr.io/websocket => github.com/coder/websocket v1.8.10
)
