package main

import (
	"os"
	"os/signal"
	"syscall"
)

var (
	signalCh   chan os.Signal
	terminated bool
)

// Terminated returns whether user has requested termination via Ctrl C,
// or other means
func Terminated() bool {
	return terminated
}

// EnableSignalHandler watches for a termination request from the user
func EnableSignalHandler() {
	if signalCh != nil {
		return
	}
	signalCh = make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGHUP, syscall.SIGINT,
		syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		s := <-signalCh
		if s != syscall.SIGQUIT {
			DebugLog.Println("Please wait while finishing up...")
		}
		terminated = true
		return
	}()
}

// TerminateSignalHandler terminates the signal handler
func TerminateSignalHandler() {
	if signalCh != nil {
		signalCh <- syscall.SIGQUIT
		close(signalCh)
	}
}
