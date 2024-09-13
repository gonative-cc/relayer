package env

import (
	"github.com/joho/godotenv"
)

// Init setups globals and env variables
func Init() error {
	if err := godotenv.Load(); err != nil {
		return err
	}
	return InitLogger()
}
