package main

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/beka-birhanu/vinom-api/udp"
	"github.com/beka-birhanu/vinom-api/udp/crypto"
	pb "github.com/beka-birhanu/vinom-api/udp/pb_encoder"
	"github.com/google/uuid"
)

type a struct{}

func (a *a) Authenticate(s []byte) (uuid.UUID, error) {
	fmt.Printf("\nAutheticated user with token %s", s)
	return uuid.New(), nil
}
func main() {
	aesKey := []byte{113, 110, 25, 53, 11, 53, 68, 33, 17, 36, 22, 7, 125, 11, 35, 16, 83, 61, 59, 49, 31, 22, 69, 17, 24, 125, 11, 35, 16, 83, 61, 59}
	asymm, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("error while generating rsa key: %s", err)
		return
	}

	serverAddr, err := net.ResolveUDPAddr("udp", "localhost:8000")
	if err != nil {
		fmt.Printf("error while resolving server addr: %s", err)
		return
	}

	rsaEnc := crypto.NewRSA(asymm)
	server, _ := udp.NewServerSocketManager(udp.ServerConfig{
		ListenAddr:    serverAddr,
		Authenticator: &a{},
		AsymmCrypto:   rsaEnc,
		SymmCrypto:    crypto.NewAESCBC(),
		Encoder:       &pb.Protobuf{},
	},
		udp.ServerWithClientRegisterHandler(func(u uuid.UUID) { fmt.Printf("\nuser %s registerd", u) }),
		udp.ServerWithReadBufferSize(2048),
		udp.ServerWithLogger(log.New(os.Stdout, "\n@Server Socket@------@", 1)),
		udp.ServerWithHeartbeatExpiration(time.Second),
	)

	client, _ := udp.NewClientServerManager(
		udp.ClientConfig{
			ServerAddr:         serverAddr,
			Encoder:            &pb.Protobuf{},
			AsymmCrypto:        crypto.NewRSA(asymm),
			ServerAsymmPubKey:  rsaEnc.GetPublicKey(),
			SymmCrypto:         crypto.NewAESCBC(),
			ClientSymmKey:      aesKey,
			AuthToken:          []byte("\"KNOCK KNOCK one\""),
			OnConnectionSucces: func() {},
			OnServerResponse: func(t byte, message []byte) {
				fmt.Printf("\n#Client Socket One#------#server responeded with: %v, %v", t, message)
			},
			OnPingResult: func(i int64) { fmt.Printf("\n#Client Socket One#------#ping result recievd: %d", i) },
		},
		udp.ClientWithLogger(log.New(os.Stdout, "\n#Client Socket One#------#", 1)),
		udp.ClientWithPingInterval(500*time.Millisecond),
	)

	client2, _ := udp.NewClientServerManager(
		udp.ClientConfig{
			ServerAddr:         serverAddr,
			Encoder:            &pb.Protobuf{},
			AsymmCrypto:        crypto.NewRSA(asymm),
			ServerAsymmPubKey:  rsaEnc.GetPublicKey(),
			SymmCrypto:         crypto.NewAESCBC(),
			ClientSymmKey:      aesKey,
			AuthToken:          []byte("\"KNOCK KNOCK two\""),
			OnConnectionSucces: func() {},
			OnServerResponse: func(t byte, message []byte) {
				fmt.Printf("\n#Client Socket Two#------#server responeded with: %v, %v", t, message)
			},
			OnPingResult: func(i int64) { fmt.Printf("\n#Client Socket Two#------#ping result recievd: %d", i) },
		},
		udp.ClientWithLogger(log.New(os.Stdout, "\n#Client Socket Two#------#", 1)),
		udp.ClientWithPingInterval(500*time.Millisecond),
	)

	go server.Serve()
	go func() {
		err = client.Connect()
		if err != nil {
			fmt.Println("unable to connect to server")
		}

	}()
	go func() {
		err = client2.Connect()
		if err != nil {
			fmt.Println("unable to connect to server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	for _ = range quit {
		server.Stop()
		client.Disconnect()
		client2.Disconnect()
		close(quit)
	}
}
