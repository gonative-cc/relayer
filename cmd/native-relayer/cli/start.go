package cli

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gonative-cc/relayer/dal"
	"github.com/gonative-cc/relayer/env"
	"github.com/gonative-cc/relayer/ika"
	"github.com/gonative-cc/relayer/ika2btc"
	"github.com/gonative-cc/relayer/native"
	"github.com/gonative-cc/relayer/native2ika"
	"github.com/gonative-cc/relayer/nbtc"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the relayer",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := prepareEnv(cmd)
		if err != nil {
			log.Error().Err(err).Msg("Failed to prepare environment")
			os.Exit(1)
		}
		db, err := initDatabase(config.DB)
		if err != nil {
			log.Error().Err(err).Msg("Failed to initialize database")
			os.Exit(1)
		}
		nativeProcessor, err := createNativeProcessor(config.Ika, db)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create Native Processor")
			os.Exit(1)
		}
		btcProcessor, err := createBTCProcessor(config.Btc, db)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create Bitcoin processor")
			os.Exit(1)
		}
		fetcher, err := createSignReqFetcher()
		if err != nil {
			log.Error().Err(err).Msg("Failed to create SignReq fetcher ")
			os.Exit(1)
		}
		relayer, err := createRelayer(
			config.Relayer,
			db,
			nativeProcessor,
			btcProcessor,
			fetcher,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create relayer")
			os.Exit(1)
		}
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := relayer.Start(cmd.Context()); err != nil {
				log.Error().Err(err).Msg("Relayer encountered an error")
			}
		}()
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		<-c
		log.Info().Msg("Stopping the relayer...")
		relayer.Stop()
		wg.Wait()
		log.Info().Msg("Relayer stopped.")
	},
}

func prepareEnv(cmd *cobra.Command) (*Config, error) {
	flags := cmd.Root().PersistentFlags()
	lvl, err := flags.GetString("log-level")
	if err != nil {
		return nil, fmt.Errorf("error getting log level: %w", err)
	}
	logLvl, err := zerolog.ParseLevel(lvl)
	if err != nil {
		return nil, fmt.Errorf("error parsing log level: %w", err)
	}
	env.InitLogger(logLvl)
	configFile, err := flags.GetString("config")
	if err != nil {
		return nil, fmt.Errorf("error getting config file path: %w", err)
	}
	config, err := loadConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}
	return config, nil
}

func initDatabase(cfg DBCfg) (*dal.DB, error) {
	db, err := dal.NewDB(cfg.File)
	if err != nil {
		return nil, fmt.Errorf("error creating database: %w", err)
	}
	if err := db.InitDB(); err != nil {
		return nil, fmt.Errorf("error initializing database: %w", err)
	}
	return db, nil
}

func createNativeProcessor(cfg IkaCfg, db *dal.DB) (*native2ika.Processor, error) {
	suiClient := sui.NewSuiClient(cfg.RPC).(*sui.Client)
	if suiClient == nil {
		return nil, fmt.Errorf("error creating Sui client")
	}
	signer, err := signer.NewSignertWithMnemonic(cfg.SignerMnemonic)
	if err != nil {
		return nil, fmt.Errorf("error creating signer with mnemonic: %w", err)
	}
	client, err := ika.NewClient(
		suiClient,
		signer,
		ika.SuiCtrCall{
			Package:  cfg.NativeLcPackage,
			Module:   cfg.NativeLcModule,
			Function: cfg.NativeLcFunction,
		},
		ika.SuiCtrCall{
			Package:  cfg.NativeLcPackage,
			Module:   cfg.NativeLcModule,
			Function: cfg.NativeLcFunction,
		},
		cfg.GasAcc,
		cfg.GasBudget,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating Ika client: %w", err)
	}
	nativeProcessor := native2ika.NewProcessor(client, db)
	return nativeProcessor, nil
}
func createBTCProcessor(btcCfg BitcoinCfg, db *dal.DB) (*ika2btc.Processor, error) {
	btcProcessor, err := ika2btc.NewProcessor(rpcclient.ConnConfig{
		Host:         btcCfg.RPCHost,
		User:         btcCfg.RPCUser,
		Pass:         btcCfg.RPCPass,
		HTTPPostMode: btcCfg.HTTPPostMode,
		DisableTLS:   btcCfg.DisableTLS,
		Params:       btcCfg.Network,
	}, btcCfg.ConfirmationThreshold, db)
	if err != nil {
		return nil, fmt.Errorf("error creating Bitcoin processor: %w", err)
	}
	return btcProcessor, nil
}

func createSignReqFetcher() (*native.APISignRequestFetcher, error) {
	// TODO: replace with the the real endpoint  once available
	return native.NewMockAPISignRequestFetcher()
}

func createRelayer(
	relayerCfg RelayerCfg,
	db *dal.DB,
	nativeProcessor *native2ika.Processor,
	btcProcessor *ika2btc.Processor,
	fetcher *native.APISignRequestFetcher,
) (*nbtc.Relayer, error) {
	relayer, err := nbtc.NewRelayer(
		nbtc.RelayerConfig(relayerCfg),
		db,
		nativeProcessor,
		btcProcessor,
		fetcher,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating relayer: %w", err)
	}
	return relayer, nil
}
