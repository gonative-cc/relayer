package main

import (
	"fmt"

	"github.com/gonative-cc/relayer/env"
	"github.com/rs/zerolog/log"
)

func main() {
	if err := env.Init(); err != nil {

		fmt.Println("\n>> Err setting up env ", err)
		return
	}

	if err := CmdExecute(); err != nil {
		log.Err(err).Msg("")
		return
	}
}
