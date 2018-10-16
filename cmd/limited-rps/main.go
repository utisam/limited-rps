package main

import (
	"os"

	"github.com/tendermint/tendermint/abci/server"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/utisam/limited-rps/lrps"
)

const (
	serverAddress = "tcp://0.0.0.0:26658"
	serverType    = "socket"
)

func main() {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))

	app := lrps.NewLRPSApplication()

	// Start the listener
	srv, err := server.NewServer(serverAddress, serverType, app)
	if err != nil {
		panic(err)
	}
	srv.SetLogger(logger.With("module", "abci-server"))
	if err := srv.Start(); err != nil {
		panic(err)
	}

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		srv.Stop()
	})

}
