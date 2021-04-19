package main

import (
	"context"
	"k8s.io/klog/v2"
	"log"
	"net"
	"os"
	"os/signal"
	"simple-distributed-cache/pkg/cache/lru"
	"simple-distributed-cache/pkg/server"
	"syscall"
	"time"
)

func mockOnEvictedFn(key string, value interface{}) error {
	log.Print("Evicted KEY:", key)
	log.Print("Evicted VALUE:", value)
	return nil
}

func main() {
	addr := "0.0.0.0:8080"
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		klog.Errorf("failed to listen on %v: %v", ln.Addr, err)
	}

	// get port
	//serverPort := ln.Addr().(*net.TCPAddr).Port
	RootCtx := SetupSignalContext()
	lruCache := lru.NewLRUCache(lru.DefaultOptions(mockOnEvictedFn))
	s := server.NewServer(addr, lruCache)
	stopCH, err := server.RunServer(RootCtx, s, ln, 10*time.Second)
	if err != nil {
		klog.Fatalf("RunServer err: %v", err)
	}
	<-stopCH
}

var onlyOneSignalHandler = make(chan struct{})
var shutdownHandler chan os.Signal
var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}

func SetupSignalContext() context.Context {
	close(onlyOneSignalHandler) // panics when called twice

	shutdownHandler = make(chan os.Signal, 2)

	ctx, cancel := context.WithCancel(context.Background())
	signal.Notify(shutdownHandler, shutdownSignals...)
	go func() {
		<-shutdownHandler
		cancel()
		<-shutdownHandler
		os.Exit(1) // second signal. Exit directly.
	}()

	return ctx
}
