package main

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peerstore"
	peer "github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"
)

func client(host host.Host, pid string, size int, reco, upload, download bool) {
	log.Println("Local peerID:", host.ID().Pretty())

	// ipfsaddr, err := ma.NewMultiaddr(dest)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	peerid, err := peer.IDB58Decode(pid)
	if err != nil {
		log.Fatalln(err)
	}

	relays := loadRelaysAddrs()

	// /ip4/51.75.127.200/udp/4141/quic/p2p/12D3KooWPwRwwKatdy5yzRVCYPHib3fntYgbFB4nqrJPHWAqXD7z/p2p-circuit
	targetAddr := ma.StringCast(fmt.Sprintf("/p2p-circuit/p2p/%s", peerid.String()))
	for _, pi := range relays {
		for _, addr := range pi.Addrs {
			maddr := ma.StringCast(fmt.Sprintf("%s/p2p/%s", addr.String(), pi.ID.String()))
			targetRelay := maddr.Encapsulate(targetAddr)
			log.Printf("added relay target: %s", targetRelay.String())
			host.Peerstore().AddAddr(peerid, targetRelay, peerstore.PermanentAddrTTL)

		}
	}

	var start time.Time
	for {
		if upload {
			start = time.Now()
			su, err := host.NewStream(context.Background(), peerid, benchUploadPID)
			if err != nil {
				log.Fatalf("New upload stream failed: %v\n", err)
			}

			reader := bufio.NewReader(su)
			if _, err = reader.ReadString('\n'); err != nil {
				log.Fatalf("Read error during stream opened ack: %v\n", err)
			}
			fmt.Printf("New upload stream took: %v\n", time.Since(start))

			data := make([]byte, size)
			rand.Read(data)

			start = time.Now()
			if _, err = su.Write(data); err != nil {
				log.Fatalf("Write error during upload: %v\n", err)
			}
			su.CloseWrite()

			if _, err = reader.ReadString('\n'); err != nil {
				log.Fatalf("Read error during uploaded ack: %v\n", err)
			}
			fmt.Printf("Data (%d bytes) upload took: %v\n", size, time.Since(start))

			su.CloseRead()
		}

		if download {
			start = time.Now()
			sd, err := host.NewStream(context.Background(), peerid, benchDownloadPID)
			if err != nil {
				log.Fatalf("New download stream failed: %v\n", err)
			}

			reader := bufio.NewReader(sd)
			if _, err = reader.ReadString('\n'); err != nil {
				log.Fatalf("Read error during stream opened ack: %v\n", err)
			}
			fmt.Printf("New download stream took: %v\n", time.Since(start))

			// Send size to download
			sizeStr := fmt.Sprintf("%d\n", size)
			if _, err = sd.Write([]byte(sizeStr)); err != nil {
				log.Fatalf("Write size error during download: %v\n", err)
			}

			start = time.Now()
			if _, err = ioutil.ReadAll(sd); err != nil {
				log.Fatalln(err)
			}
			fmt.Printf("Data (%d bytes) download took: %v\n", size, time.Since(start))

			sd.Close()
		}

		if reco {
			reco = false
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Reconnection test: switch connection then press enter")
			_, _ = reader.ReadString('\n')
			continue
		}

		break
	}
}
