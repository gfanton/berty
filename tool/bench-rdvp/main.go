package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/oklog/run"
	"github.com/peterbourgon/ff/v3/ffcli"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	p2p_rpdb "github.com/libp2p/go-libp2p-rendezvous/db/sqlite"
	ma "github.com/multiformats/go-multiaddr"
)

// global flags
var (
	logger = zap.NewNop()

	// root flags
	rootFlagSet = flag.NewFlagSet("bench", flag.ExitOnError)
	verbose     = rootFlagSet.Bool("v", false, "increase log verbosity")

	// serve flags
	serveFlagSet   = flag.NewFlagSet("serve", flag.ExitOnError)
	serveListeners = serveFlagSet.String("l", "", "swarm listeners")
)

func serveCmd(ctx context.Context, args []string) error {
	db, err := p2p_rpdb.OpenDB(ctx, ":memory:")
	if err != nil {
		return fmt.Errorf("unable to open db: %w", err)
	}

	defer db.Close()

	opts := RDVPOpts{
		Logger: logger,
		DB:     db,
	}

	if *serveListeners != "" {
		listeners := strings.Split(*serveListeners, ",")
		opts.Listeners = make([]ma.Multiaddr, len(listeners))
		for i, a := range listeners {
			opts.Listeners[i], err = ma.NewMultiaddr(a)
			if err != nil {
				return fmt.Errorf("unable to parse maddr(%s): %w", a, err)
			}
		}
	}

	rdvp, err := NewRDVPHost(ctx, &opts)
	if err != nil {
		return fmt.Errorf("unable to create rdvp: %w", err)
	}

	defer rdvp.Close()

	maddrs := rdvp.Addrs()
	addrs := make([]string, len(maddrs))
	for i, m := range maddrs {
		addrs[i] = m.String()
	}

	logger.Info("serving host", zap.Strings("addrs", addrs))

	// wait for context
	<-ctx.Done()
	return ctx.Err()
}

func main() {
	server := &ffcli.Command{
		Name:        "serve",
		ShortUsage:  "bench server",
		ShortHelp:   "serve rdvp",
		FlagSet:     serveFlagSet,
		Subcommands: []*ffcli.Command{},
		Exec:        serveCmd,
	}

	root := &ffcli.Command{
		Name:    "bench [flags] <subcommand>",
		FlagSet: rootFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
		Subcommands: []*ffcli.Command{
			server,
		},
	}

	args := os.Args[1:]
	if err := root.Parse(args); err != nil {
		log.Fatalf("unable to parse flags: %s", err)
	}

	// init logger
	logger = initLogger(*verbose)
	defer logger.Sync()

	// create process context
	processCtx, processCancel := context.WithCancel(context.Background())
	var process run.Group
	{
		// handle interrupt signals
		execute, interrupt := run.SignalHandler(processCtx, os.Interrupt)
		process.Add(execute, interrupt)

		// add root command to process
		process.Add(func() error {
			return root.Run(processCtx)
		}, func(error) {
			processCancel()
		})
	}

	// start process
	switch err := process.Run(); err {
	case flag.ErrHelp, nil: // ok
	case context.Canceled, context.DeadlineExceeded:
		logger.Fatal("interrupted", zap.Error(err))
	default:
		logger.Fatal(err.Error())
	}
}

func initLogger(verbose bool) *zap.Logger {
	var level zapcore.Level
	if verbose {
		level = zapcore.DebugLevel
	} else {
		level = zapcore.InfoLevel
	}

	encodeConfig := zap.NewDevelopmentEncoderConfig()
	encodeConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encodeConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.Stamp)
	consoleEncoder := zapcore.NewConsoleEncoder(encodeConfig)
	consoleDebugging := zapcore.Lock(os.Stdout)
	core := zapcore.NewCore(consoleEncoder, consoleDebugging, level)
	logger := zap.New(core)

	logger.Debug("logger initialised")
	return logger
}
