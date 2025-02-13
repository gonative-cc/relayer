package main

import (
	"github.com/rs/zerolog/log"
)

func main() {
	if err := CmdExecute(); err != nil {
		log.Err(err).Msg("")
		return
	}
}
