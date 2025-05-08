package suigoclient

import (
	"github.com/pattonkan/sui-go/sui"
	"github.com/pattonkan/sui-go/sui/suiptb"
	"github.com/pattonkan/sui-go/suiclient"
	"github.com/pattonkan/sui-go/suisigner"
	"github.com/rs/zerolog"
)

type BTCLCObject struct {
	logger      zerolog.Logger
	suiClient   *suiclient.ClientImpl
	signer      *suisigner.Signer
	lcObjectID  *sui.ObjectId
	lcPackageID *sui.PackageId
}
