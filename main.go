package main

import (
	"context"
	"flag"
	"k8s.io/klog/v2"
	"log"
	"net"
	"os"
	"os/signal"
	"simple-distributed-cache/pkg/cache/lru"
	"simple-distributed-cache/pkg/server"
	"strconv"
	"syscall"
	"time"
)

func mockOnEvictedFn(key string, value interface{}) error {
	log.Print("Evicted KEY:", key)
	log.Print("Evicted VALUE:", value)
	return nil
}

func main() {

	var hostName string
	var port int
	var proxy bool
	var verbosity int
	flag.StringVar(&hostName, "hostname", "0.0.0.0", "server host name")
	flag.IntVar(&port, "port", 8080, "server port")
	flag.BoolVar(&proxy, "proxy", true, "Is a proxy server?")
	flag.IntVar(&verbosity, "v", 5, "number for the log level verbosity")
	flag.Parse()

	addr := net.JoinHostPort(hostName, strconv.Itoa(port))

	ln, err := net.Listen("tcp", addr)
	var level klog.Level
	level.Set(strconv.Itoa(verbosity))
	defer klog.Flush()
	if err != nil {
		klog.Errorf("failed to listen on %v: %v", ln.Addr, err)
	}
	peerServer := []string{
		"0.0.0.0:8081",
		"0.0.0.0:8082",
		"0.0.0.0:8083",
	}

	// get port
	//serverPort := ln.Addr().(*net.TCPAddr).Port
	rootCtx := SetupSignalContext()
	lruCache := lru.NewLRUCache(lru.DefaultOptions(mockOnEvictedFn))

	s := server.NewServer(addr, lruCache)
	s.Set(nil, peerServer...)

	stopCH, err := server.RunServer(rootCtx, s, ln, 10*time.Second)
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
