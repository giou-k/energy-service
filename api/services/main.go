package main

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/ardanlabs/conf/v3"

	"github.com/energy-service/api/services/debug"
	"github.com/energy-service/api/services/route"
	"github.com/energy-service/platform/logger"
)

var build = "develop"

func main() {
	log := initLogger()
	// -------------------------------------------------------------------------

	ctx := context.Background()

	if err := run(ctx, log); err != nil {
		log.Error(ctx, "startup", "err", err)
		os.Exit(1)
	}
}

func initLogger() *logger.Logger {
	var log *logger.Logger

	events := logger.Events{
		Error: func(ctx context.Context, r logger.Record) {
			log.Info(ctx, "******* SEND ALERT *******")
		},
	}

	traceIDFn := func(ctx context.Context) string {
		return ""
	}

	log = logger.NewWithEvents(os.Stdout, logger.LevelInfo, "ENERGY", traceIDFn, events)

	return log
}

func run(ctx context.Context, log *logger.Logger) error {
	// -------------------------------------------------------------------------
	// GOMAXPROCS

	log.Info(ctx, "startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	// -------------------------------------------------------------------------
	// Configuration

	cfg := struct {
		conf.Version
		Web struct {
			ReadTimeout        time.Duration `conf:"default:5s"`
			WriteTimeout       time.Duration `conf:"default:10s"`
			IdleTimeout        time.Duration `conf:"default:120s"`
			ShutdownTimeout    time.Duration `conf:"default:20s"`
			APIHost            string        `conf:"default:0.0.0.0:3000"`
			DebugHost          string        `conf:"default:0.0.0.0:3011"`
			CORSAllowedOrigins []string      `conf:"default:*, mask"`
		}
	}{
		Version: conf.Version{
			Build: build,
			Desc:  "Energy",
		},
	}

	const prefix = "ENERGY"
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	// -------------------------------------------------------------------------
	// App Starting

	log.Info(ctx, "starting service", "version", cfg.Build)
	defer log.Info(ctx, "shutdown complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	log.Info(ctx, "startup", "config", out)

	log.BuildInfo(ctx)

	expvar.NewString("build").Set(cfg.Build)

	// -------------------------------------------------------------------------
	// Start Debug Service
	go func() {
		log.Info(ctx, "startup", "status", "debug v1 router started", "host", cfg.Web.DebugHost)

		//  note that this goroutine is an orphan of the main go routine, but we don't need to end properly this debug goroutine,
		// since it is just writing the state of the app, and when the app ends, it can end too.
		if err = http.ListenAndServe(cfg.Web.DebugHost, debug.Router()); err != nil {
			log.Info(ctx, "shutdown", "status", "debug v1 router closed", "host", cfg.Web.DebugHost, "msg", err)
		}
	}()

	// -------------------------------------------------------------------------
	// Start API Service

	log.Info(ctx, "startup", "status", "initializing V1 API support")

	shutdown := make(chan os.Signal, 1)
	// Notify our shutdown channel for the following signals.
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	srv := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      route.WebAPI(),
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
		ErrorLog:     logger.NewStdLogger(log, logger.LevelError),
	}

	serverError := make(chan error, 1)
	go func() {
		log.Info(ctx, "startup", "status", "api router started", "host", srv.Addr)
		serverError <- srv.ListenAndServe()
	}()

	// -------------------------------------------------------------------------
	// Shutdown

	select {
	case srvErr := <-serverError:
		return fmt.Errorf("server error: %w", srvErr)
	case sig := <-shutdown:
		log.Info(ctx, "shutdown", "status", "shutdown started", "signal", sig.String())
		defer log.Info(ctx, "shutdown", "status", "shutdown complete", "signal", sig.String())

		ctx, cancel := context.WithTimeout(ctx, cfg.Web.ShutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			srv.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}

	}

	return nil
}
