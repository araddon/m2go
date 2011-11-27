package m2go

import (
  "os"
  "os/signal"
)

// -----------------------------------------------------------------------------
// Signal Handling  from  https://github.com/tav/ampify/blob/master/src/amp/runtime/runtime.go
// -----------------------------------------------------------------------------

var signalHandlers = make(map[os.UnixSignal]func())

func RegisterSignalHandler(signal os.UnixSignal, handler func()) {
  signalHandlers[signal] = handler
}

func ClearSignalHandler(signal os.UnixSignal) {
  signalHandlers[signal] = func() {}
}

func handleSignals() {
  var sig os.Signal
  for {
    sig = <-signal.Incoming
    handler, found := signalHandlers[sig.(os.UnixSignal)]
    if found {
      handler()
    }
  }
}

var exitHandlers = []func(){}

func RunExitHandlers() {
  for _, handler := range exitHandlers {
    handler()
  }
}

func RegisterExitHandler(handler func()) {
  exitHandlers = append(exitHandlers, handler)
}

func Exit(code int) {
  RunExitHandlers()
  os.Exit(code)
}

