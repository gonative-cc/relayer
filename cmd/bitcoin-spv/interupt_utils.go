package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func initalize() {
	signals = []os.Signal{os.Interrupt, syscall.SIGTERM}
}

var (
	interruptChannel         chan os.Signal
	addHandlerChannel        = make(chan func())
	interruptHandlersDone    = make(chan struct{})
	simulateInterruptChannel = make(chan struct{}, 1)
	signals                  = []os.Signal{os.Interrupt}
)

// mainInterruptHandler listens for SIGINT (Ctrl+C) signals on the
// interruptChannel and invokes the registered interruptCallbacks accordingly.
// It also listens for callback registration.
func mainInterruptHandler() {
	var handlers []func()

	// invokeCallbacks runs all registered handlers in LIFO order
	executeHandlers := func() {
		for i := len(handlers) - 1; i >= 0; i-- {
			handlers[i]()
		}
		close(interruptHandlersDone)
	}

	// handleShutdown processes shutdown requests from different sources
	processShutdown := func(msg string) {
		fmt.Print(msg)
		executeHandlers()
	}

mainLoop:
	for {
		select {
		case sig := <-interruptChannel:
			processShutdown(fmt.Sprintf("Received signal (%s). Shutting down...", sig))
			break mainLoop
		case <-simulateInterruptChannel:
			processShutdown("Received shutdown request. Shutting down...")
			break mainLoop
		case h := <-addHandlerChannel:
			handlers = append(handlers, h)
		}
	}
}

// addHandler registers a new interrupt handler function.
// It initializes the interrupt handling system if this is the first handler.
func addHandler(callback func()) {
	// Initialize interrupt handling if not already done
	if interruptChannel == nil {
		interruptChannel = make(chan os.Signal, 1)
		signal.Notify(interruptChannel, signals...)
		go mainInterruptHandler()
	}

	// Register the new callback handler
	addHandlerChannel <- callback
}
