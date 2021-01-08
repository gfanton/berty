package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	ma "github.com/multiformats/go-multiaddr"
)

func server(host host.Host, quic, insecure bool) {
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", host.ID().Pretty()))
	addr := host.Addrs()[0]
	fullAddr := addr.Encapsulate(hostAddr)

	log.Printf("Server MultiAddr: %s\n", fullAddr)

	hint := fmt.Sprintf("Now run this cmd with flags '-dest %s", fullAddr)
	if insecure {
		hint += " -insecure"
	}
	if quic {
		hint += " -quic"
	}
	hint += "'\n"
	log.Println(hint)

	host.SetStreamHandler(benchDownloadPID, func(s network.Stream) {
		defer s.Close()

		remotePeerID := s.Conn().RemotePeer().Pretty()
		log.Printf("New download stream from: %s\n", remotePeerID)

		_, err := s.Write([]byte("\n"))
		if err != nil {
			log.Printf("Write error during stream opened ack to %s: %v\n", remotePeerID, err)
		}

		buf := bufio.NewReader(s)
		str, err := buf.ReadString('\n')
		if err != nil {
			log.Printf("Read error during download from %s: %v\n", remotePeerID, err)
			return
		}

		size, err := strconv.Atoi(string(str[:len(str)-1]))
		if err != nil {
			log.Printf("Invalid size received: %s: %v", str, err)
			return
		}

		data := make([]byte, size)
		rand.Read(data)

		_, err = s.Write(data)
		if err != nil {
			log.Printf("Write error during download to %s: %v\n", remotePeerID, err)
		}
		log.Printf("Sent %d bytes to %s", len(data), remotePeerID)
	})

	host.SetStreamHandler(benchUploadPID, func(s network.Stream) {
		defer s.Close()

		remotePeerID := s.Conn().RemotePeer().Pretty()
		log.Printf("New upload stream from: %s\n", remotePeerID)

		_, err := s.Write([]byte("\n"))
		if err != nil {
			log.Printf("Write error during stream opened ack to %s: %v\n", remotePeerID, err)
		}

		reader := bufio.NewReader(s)
		data, err := ioutil.ReadAll(reader)
		if err != nil {
			log.Printf("Read error during upload from %s: %v\n", remotePeerID, err)
			return
		}
		log.Printf("Received %d bytes from %s", len(data), remotePeerID)

		_, err = s.Write([]byte("\n"))
		if err != nil {
			log.Printf("Write error during uploaded ack to %s: %v\n", remotePeerID, err)
		}
	})

	go func() {
		for {
			time.Sleep(time.Second)
			fmt.Println(host.Addrs())
		}
	}()

	select {}
}
