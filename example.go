// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ingore
// +build ingore

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lib-x/winsvc"
)

var (
	serviceName        = flag.String("name", "example-service", "Service name")
	serviceDisplayName = flag.String("display", "Example Windows Service", "Service display name")
	serviceDescription = flag.String("desc", "An example Windows service", "Service description")
	serviceInstall     = flag.Bool("install", false, "Install the service")
	serviceUninstall   = flag.Bool("uninstall", false, "Uninstall the service")
	serviceStart       = flag.Bool("start", false, "Start the service")
	serviceStop        = flag.Bool("stop", false, "Stop the service")
)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run() error {
	if *serviceInstall {
		return installService()
	}
	if *serviceUninstall {
		return winsvc.RemoveService(*serviceName)
	}
	if *serviceStart {
		return winsvc.StartService(*serviceName)
	}
	if *serviceStop {
		return winsvc.StopService(*serviceName)
	}

	if winsvc.InServiceMode() {
		return winsvc.RunAsService(*serviceName, startServer, stopServer, false)
	}

	return startServer()
}

func installService() error {
	exePath, err := winsvc.GetAppPath()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	options := []winsvc.ServiceOption{
		winsvc.DisplayName(*serviceDisplayName),
		winsvc.Description(*serviceDescription),
		winsvc.AutoStart(),
	}

	return winsvc.InstallServiceWithOption(exePath, *serviceName, nil, options...)
}

func startServer() error {
	srv := &http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Example service running at %s", time.Now().Format(time.RFC3339))
		}),
	}

	go func() {
		log.Printf("Server starting on http://localhost%s", srv.Addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	return stopServer()
}

func stopServer() error {
	log.Println("Server shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := http.DefaultServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	log.Println("Server stopped")
	return nil
}
