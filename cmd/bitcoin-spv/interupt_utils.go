package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	interruptSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}
}

var (
	interruptChan       chan os.Signal
	registerHandlerChan = make(chan func())
	interruptDone       = make(chan struct{})
	simulateChan        = make(chan struct{}, 1)
	interruptSignals    = []os.Signal{os.Interrupt}
)

// mainInterruptHandler manages system interrupt signals and callback execution.
// It processes SIGINT/SIGTERM signals and executes registered handlers in reverse order.
// New handlers can be dynamically added through a dedicated channel.
func mainInterruptHandler() {
	var handlers []func()

	// invokeCallbacks runs all registered handlers in LIFO order
	executeHandlers := func() {
		for i := len(handlers) - 1; i >= 0; i-- {
			handlers[i]()
		}
		close(interruptDone)
	}

	// processShutdown handles system termination signals from various inputs
	processShutdown := func(msg string) {
		fmt.Print(msg)
		executeHandlers()
	}

mainLoop:
	for {
		select {
		case sig := <-interruptChan:
			processShutdown(fmt.Sprintf("Signal %s detected. Initiating shutdown...", sig))
			break mainLoop
		case <-simulateChan:
			processShutdown("Shutdown request received. System going down...")
			break mainLoop
		case h := <-registerHandlerChan:
			handlers = append(handlers, h)
		}
	}
}

// registerHandler adds a function to be called when system interruption occurs.
// On first registration, it sets up the interrupt handling infrastructure.
func registerHandler(callback func()) {
	// Initialize interrupt handling if not already done
	if interruptChan == nil {
		interruptChan = make(chan os.Signal, 1)
		signal.Notify(interruptChan, interruptSignals...)
		go mainInterruptHandler()
	}

	// Register the new callback handler
	registerHandlerChan <- callback
}
