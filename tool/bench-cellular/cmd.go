package main

import (
	"flag"
	"log"
	"os"

	golog "github.com/ipfs/go-log"
)

const (
	benchDownloadPID = "/bench/download/1.0.0"
	benchUploadPID   = "/bench/upload/1.0.0"
)

func main() {
	dest := flag.String("dest", "", "(client) server peer to dial")
	reco := flag.Bool("reco", false, "(client) test reconnection to server")
	upload := flag.Bool("upload", false, "(client) upload data to server")
	download := flag.Bool("download", false, "(client) download data from server")
	size := flag.Int("size", 0, "(client) size of data to upload/download")
	quic := flag.Bool("quic", false, "use QUIC instead of TCP")
	autorelay := flag.Bool("autorelay", false, "use AutoRelay instead of Berty's hardcoded relays")
	ip6 := flag.Bool("ip6", false, "use ipv6 instead of ipv4")
	port := flag.Int("port", 0, "port to listen on (default: random)")
	insecure := flag.Bool("insecure", false, "use an unencrypted connection")
	seed := flag.Int64("seed", 0, "set random seed for id generation")
	flag.Parse()

	if *dest != "" && !*upload && !*download {
		log.Print("client must pass -upload, -download or both\n\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *dest != "" && *size <= 0 {
		log.Print("client must pass -size\n\n")
		flag.PrintDefaults()
		os.Exit(2)
	}
	if *dest == "" && *reco {
		log.Print("reco flag is for client only\n\n")
		flag.PrintDefaults()
		os.Exit(3)
	}

	// TODO: display relevant logs only at debug level
	golog.SetAllLoggers(golog.LevelDebug)
	// golog.SetAllLoggers(golog.LevelError)

	host, err := createBasicHost(*seed, *port, *autorelay, *insecure, *quic, *ip6)
	if err != nil {
		log.Fatal(err)
	}

	if *dest == "" {
		server(host, *quic, *insecure)
	}
	client(host, *dest, *size, *reco, *upload, *download)
}
