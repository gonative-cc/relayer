package cli

import (
	"errors"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops the running relayer",
	Run: func(cmd *cobra.Command, args []string) {
		stopRealyer()
	},
}

func stopRealyer() {
	_, err := os.Stat(PidFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Info().Msg("Realyer is not running")
			return
		}
		log.Error().Err(err).Msg("Failed checking for PID file")
		return
	}

	fileContent, err := os.ReadFile(PidFilePath)
	if err != nil {
		log.Error().Err(err).Msg("Failed reading PID file")
		return
	}

	pid, err := strconv.Atoi(string(fileContent))
	if err != nil {
		log.Error().Err(err).Msgf("Invalid PID found in %s", PidFilePath)
		os.Remove(PidFilePath)
		return
	}

	// Send SIGTERM
	err = syscall.Kill(pid, syscall.SIGTERM)
	if err != nil {
		log.Error().Err(err).Msg("Realyer process not found")
		os.Remove(PidFilePath) // Clean up
		return
	}
	log.Info().Msgf("Sent SIGTERM to process %d", pid)

	for i := 0; i < 5; i++ { // we use loop to wait up to 5 seconds for the SIGTERM signal to be processed
		time.Sleep(1 * time.Second)
		_, err := os.Stat(PidFilePath)
		if errors.Is(err, os.ErrNotExist) {
			log.Info().Msg("Realyer stopped successfully.")
			return
		} else if err != nil {
			log.Error().Err(err).Msg("Error while checking file")
			return
		}
	}
	log.Warn().Msg("Realyer did not stop within the timeout")
}
