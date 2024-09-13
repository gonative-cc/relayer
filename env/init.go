package env

import (
	"fmt"

	"github.com/joho/godotenv"
)

// Init setups globals and env variables
func Init() error {
	if err := godotenv.Load(); err != nil {
		fmt.Println("\n>> err loading ENV ", err.Error())
		return err
	}
	return InitLogger()
}
