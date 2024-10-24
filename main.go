package main

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

//var build = "develop"

func main() {
	var log *logger.Logger

	events := logger.Events{
		Error: func(ctx context.Context, r logger.Record) {
			log.Info(ctx, "******* SEND ALERT *******")
		},
	}

	traceIDFn := func(ctx context.Context) string {
		return ""
	}

	log = logger.NewWithEvents(os.Stdout, logger.LevelInfo, "SALES", traceIDFn, events)

	// -------------------------------------------------------------------------

	ctx := context.Background()

	if err := run(ctx, log); err != nil {
		log.Error(ctx, "startup", "err", err)
		return // Adding this return just to be on the safe side that the program end here, even if someone adds more code after that clause.
	}
	// ----------------------------------------
	log.Println("starting service", build)
	defer log.Println("service ended")

	log.Println("startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	shutdown := make(chan os.Signal, 1)
	// Notify our shutdown channel for the following signals.
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	log.Println("stopping service")
}

func run(ctx context.Context, log *logger.Logger) error {
	log.Println("starting service", build)
	defer log.Println("service ended")

	log.Println("startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))
	shutdown := make(chan os.Signal, 1)
	// Notify our shutdown channel for the following signals.
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	log.Info(ctx, "shutdown", "status", "shutdown started", "signal", sig)
	defer log.Info(ctx, "shutdown", "status", "shutdown complete", "signal", sig)

	return nil
}
