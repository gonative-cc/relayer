package main

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/gonative-cc/relayer/env"
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
