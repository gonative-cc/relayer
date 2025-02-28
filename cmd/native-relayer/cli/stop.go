package cli

import (
	"errors"
	"fmt"
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
		if err := stopRealyer(); err != nil {
			log.Error().Err(err).Msg("Failed to stop relayer")
			os.Exit(1)
		}
	},
}

func stopRealyer() error {
	_, err := os.Stat(PidFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Info().Msg("Realyer is not running")
			return nil
		}
		return fmt.Errorf("check for PID file: %w", err)
	}

	fileContent, err := os.ReadFile(PidFilePath)
	if err != nil {
		return fmt.Errorf("read PID file: %w", err)
	}

	pid, err := strconv.Atoi(string(fileContent))
	if err != nil {
		os.Remove(PidFilePath)
		return fmt.Errorf("invalid PID found: %w", err)
	}

	// Send SIGTERM
	err = syscall.Kill(pid, syscall.SIGTERM)
	if err != nil {
		os.Remove(PidFilePath) // Clean up
		return fmt.Errorf("kill process: %w", err)
	}
	log.Info().Msgf("Sent SIGTERM to process %d", pid)

	for i := 0; i < 10; i++ { // we use loop to wait for the SIGTERM signal to be processed
		time.Sleep(1 * time.Second)
		_, err := os.Stat(PidFilePath)
		if errors.Is(err, os.ErrNotExist) {
			log.Info().Msg("Realyer stopped successfully.")
			return nil
		} else if err != nil {
			return fmt.Errorf("check file after SIGTERM: %w", err)
		}
	}
	log.Warn().Msg("Realyer did not stop within the timeout")
	return nil
}
