package main

import (
	"context"
	crand "crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	mrand "math/rand"
	"net"
	"net/http"
	"os"
	"strings"

	// nolint:staticcheck
	libp2p "github.com/libp2p/go-libp2p"
	libp2p_cicuit "github.com/libp2p/go-libp2p-circuit"
	libp2p_ci "github.com/libp2p/go-libp2p-core/crypto"
	libp2p_host "github.com/libp2p/go-libp2p-core/host"
	metrics "github.com/libp2p/go-libp2p-core/metrics"
	libp2p_peer "github.com/libp2p/go-libp2p-core/peer"
	libp2p_quic "github.com/libp2p/go-libp2p-quic-transport"
	libp2p_rp "github.com/libp2p/go-libp2p-rendezvous"
	libp2p_rpdb "github.com/libp2p/go-libp2p-rendezvous/db/sqlite"
	"github.com/libp2p/go-libp2p/config"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/oklog/run"
	ff "github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"moul.io/srand"

	"berty.tech/berty/v2/go/internal/ipfsutil"
	"berty.tech/berty/v2/go/internal/logutil"
	"berty.tech/berty/v2/go/pkg/errcode"
)

func main() {
	log.SetFlags(0)

	// opts
	var (
		logFormat             = "color"   // json, console, color, light-console, light-color
		logToFile             = "stderr"  // can be stdout, stderr or a file path
		logFilters            = "info+:*" // info and more for everythign
		serveURN              = ":memory:"
		serveListeners        = "/ip4/0.0.0.0/tcp/4040,/ip4/0.0.0.0/udp/4141/quic"
		servePK               = ""
		serveAnnounce         = ""
		serveMetricsListeners = ""
		genkeyType            = "Ed25519"
		genkeyLength          = 2048
	)

	// parse opts
	var (
		globalFlags = flag.NewFlagSet("berty", flag.ExitOnError)
		serveFlags  = flag.NewFlagSet("serve", flag.ExitOnError)
		genkeyFlags = flag.NewFlagSet("genkey", flag.ExitOnError)
	)
	globalFlags.StringVar(&logFilters, "log.filters", logFilters, "logged namespaces")
	globalFlags.StringVar(&logToFile, "log.file", logToFile, "if specified, will log everything in JSON into a file and nothing on stderr")
	globalFlags.StringVar(&logFormat, "log.format", logFormat, "if specified, will override default log format")
	serveFlags.StringVar(&serveURN, "db", serveURN, "rdvp sqlite URN")
	serveFlags.StringVar(&serveListeners, "l", serveListeners, "lists of listeners of (m)addrs separate by a comma")
	serveFlags.StringVar(&servePK, "pk", servePK, "private key (generated by `rdvp genkey`)")
	serveFlags.StringVar(&serveAnnounce, "announce", serveAnnounce, "addrs that will be announce by this server")
	serveFlags.StringVar(&serveMetricsListeners, "metrics", serveMetricsListeners, "metrics listener, if empty will disable metrics")
	genkeyFlags.StringVar(&genkeyType, "type", genkeyType, "Type of the private key generated, one of : Ed25519, ECDSA, Secp256k1, RSA")
	genkeyFlags.IntVar(&genkeyLength, "length", genkeyLength, "The length (in bits) of the key generated.")

	serve := &ffcli.Command{
		Name:       "serve",
		ShortUsage: "rdvp [global flags] serve [flags]",
		LongHelp:   "EXAMPLE\n  rdvp genkey > rdvp.key\n  rdvp serve -pk `cat rdvp.key` -db ./rdvp-store",
		FlagSet:    serveFlags,
		Options:    []ff.Option{ff.WithEnvVarPrefix("RDVP")},
		Exec: func(ctx context.Context, args []string) error {
			if len(args) > 0 {
				return flag.ErrHelp
			}

			mrand.Seed(srand.MustSecure())
			logger, cleanup, err := logutil.NewLogger(logFilters, logFormat, logToFile)
			if err != nil {
				return errcode.TODO.Wrap(err)
			}
			defer cleanup()

			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			var gServe run.Group
			gServe.Add(func() error {
				<-ctx.Done()
				return ctx.Err()
			}, func(error) {
				cancel()
			})

			laddrs := strings.Split(serveListeners, ",")
			listeners, err := ipfsutil.ParseAddrs(laddrs...)
			if err != nil {
				return errcode.TODO.Wrap(err)
			}

			// load existing or generate new identity
			var priv libp2p_ci.PrivKey
			if servePK != "" {
				kBytes, err := base64.StdEncoding.DecodeString(servePK)
				if err != nil {
					return errcode.TODO.Wrap(err)
				}
				priv, err = libp2p_ci.UnmarshalPrivateKey(kBytes)
				if err != nil {
					return errcode.TODO.Wrap(err)
				}
			} else {
				// Don't use key params here, this is a dev tool, a real installation should use a static key.
				priv, _, err = libp2p_ci.GenerateKeyPairWithReader(libp2p_ci.Ed25519, -1, crand.Reader) // nolint:staticcheck
				if err != nil {
					return errcode.TODO.Wrap(err)
				}
			}

			var addrsFactory config.AddrsFactory = func(ms []ma.Multiaddr) []ma.Multiaddr { return ms }
			if serveAnnounce != "" {
				aaddrs := strings.Split(serveAnnounce, ",")
				announces, err := ipfsutil.ParseAddrs(aaddrs...)
				if err != nil {
					return errcode.TODO.Wrap(err)
				}

				addrsFactory = func([]ma.Multiaddr) []ma.Multiaddr { return announces }
			}

			reporter := metrics.NewBandwidthCounter()

			// init p2p host
			host, err := libp2p.New(ctx,
				// default tpt + quic
				libp2p.DefaultTransports,
				libp2p.Transport(libp2p_quic.NewTransport),

				// Nat & Relay service
				libp2p.EnableNATService(),
				libp2p.DefaultStaticRelays(),
				libp2p.EnableRelay(libp2p_cicuit.OptHop),

				// swarm listeners
				libp2p.ListenAddrs(listeners...),

				// identity
				libp2p.Identity(priv),

				// announce
				libp2p.AddrsFactory(addrsFactory),

				// metrics
				libp2p.BandwidthReporter(reporter),
			)
			if err != nil {
				return errcode.TODO.Wrap(err)
			}

			defer host.Close()
			logHostInfo(logger, host)

			db, err := libp2p_rpdb.OpenDB(ctx, serveURN)
			if err != nil {
				return errcode.TODO.Wrap(err)
			}

			defer db.Close()

			// start service
			_ = libp2p_rp.NewRendezvousService(host, db)

			if serveMetricsListeners != "" {
				ml, err := net.Listen("tcp", serveMetricsListeners)
				if err != nil {
					return errcode.TODO.Wrap(err)
				}

				registery := prometheus.NewRegistry()
				registery.MustRegister(prometheus.NewBuildInfoCollector())
				registery.MustRegister(prometheus.NewGoCollector())
				registery.MustRegister(ipfsutil.NewHostCollector(host))
				registery.MustRegister(ipfsutil.NewBandwidthCollector(reporter))
				// @TODO(gfanton): add rdvp specific collector...

				handerfor := promhttp.HandlerFor(
					registery,
					promhttp.HandlerOpts{Registry: registery},
				)

				mux := http.NewServeMux()
				gServe.Add(func() error {
					mux.Handle("/metrics", handerfor)
					logger.Info("metrics listener",
						zap.String("handler", "/metrics"),
						zap.String("listener", ml.Addr().String()))
					return http.Serve(ml, mux)
				}, func(error) {
					ml.Close()
				})

			}

			if err = gServe.Run(); err != nil {
				return errcode.TODO.Wrap(err)
			}
			return nil
		},
	}

	genkey := &ffcli.Command{
		Name:    "genkey",
		FlagSet: genkeyFlags,
		Exec: func(context.Context, []string) error {
			keyType, ok := keyNameToKeyType[strings.ToLower(genkeyType)]
			if !ok {
				return fmt.Errorf("unknown key type : '%s'. Only Ed25519, ECDSA, Secp256k1, RSA supported", genkeyType)
			}
			priv, _, err := libp2p_ci.GenerateKeyPairWithReader(keyType, genkeyLength, crand.Reader) // nolint:staticcheck
			if err != nil {
				return errcode.TODO.Wrap(err)
			}

			kBytes, err := libp2p_ci.MarshalPrivateKey(priv)
			if err != nil {
				return errcode.TODO.Wrap(err)
			}

			fmt.Println(base64.StdEncoding.EncodeToString(kBytes))
			return nil
		},
	}

	root := &ffcli.Command{
		ShortUsage:  "rdvp [global flags] <subcommand>",
		FlagSet:     globalFlags,
		Options:     []ff.Option{ff.WithEnvVarPrefix("RDVP")},
		Subcommands: []*ffcli.Command{serve, genkey},
		Exec: func(context.Context, []string) error {
			return flag.ErrHelp
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var process run.Group
	// handle close signal
	execute, interrupt := run.SignalHandler(ctx, os.Interrupt)
	process.Add(execute, interrupt)

	// add root command to process
	process.Add(func() error {
		return root.ParseAndRun(ctx, os.Args[1:])
	}, func(error) {
		cancel()
	})

	// run process
	if err := process.Run(); err != nil && err != context.Canceled {
		log.Println(err)
		return
	}
}

// Names are in lower case.
var keyNameToKeyType = map[string]int{
	"ed25519":   libp2p_ci.Ed25519,
	"ecdsa":     libp2p_ci.ECDSA,
	"secp256k1": libp2p_ci.Secp256k1,
	"rsa":       libp2p_ci.RSA,
}

// helpers

func logHostInfo(l *zap.Logger, host libp2p_host.Host) {
	// print peer addrs
	fields := []zapcore.Field{
		zap.String("host ID (local)", host.ID().String()),
	}

	addrs := host.Addrs()
	pi := libp2p_peer.AddrInfo{
		ID:    host.ID(),
		Addrs: addrs,
	}
	if maddrs, err := libp2p_peer.AddrInfoToP2pAddrs(&pi); err == nil {
		for _, maddr := range maddrs {
			fields = append(fields, zap.Stringer("maddr", maddr))
		}
	}

	l.Info("host started", fields...)
}
