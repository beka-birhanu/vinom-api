package main

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/beka-birhanu/vinom-api/infrastruture/crypto"
	gamepb "github.com/beka-birhanu/vinom-api/infrastruture/pb_encoder/game"
	udppb "github.com/beka-birhanu/vinom-api/infrastruture/pb_encoder/udp"
	"github.com/beka-birhanu/vinom-api/infrastruture/udp"
	maze "github.com/beka-birhanu/vinom-api/infrastruture/willson_maze"
	game "github.com/beka-birhanu/vinom-api/service"
	"github.com/beka-birhanu/vinom-api/service/i"
	"github.com/google/uuid"
)

type a struct{}

func (a *a) Authenticate(s []byte) (uuid.UUID, error) {
	return uuid.FromBytes(s)
}

func main1() {
	maz, _ := maze.New(10, 10)
	if maze.PopulateReward(maze.RewardModel{RewardOne: 1, RewardTwo: 5, RewardTypeProb: 0.9}, maz) != nil {
		return
	}

	game_encoder := &gamepb.Protobuf{}

	p2pos := game_encoder.NewCellPosition()
	p2pos.SetCol(1)
	p2pos.SetRow(1)
	player2 := game_encoder.NewPlayer()
	player2.SetID(uuid.New())
	player2.SetPos(p2pos)

	p1pos := game_encoder.NewCellPosition()
	p1pos.SetCol(1)
	p1pos.SetRow(2)
	player1 := game_encoder.NewPlayer()
	player1.SetID(uuid.New())
	player1.SetPos(p1pos)
	if p1pos == nil {
		return
	}
	gameServer, err := game.NewGame(maz, []i.Player{player1, player2}, game_encoder)
	if err != nil {
		return
	}

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
		Encoder:       &udppb.Protobuf{},
		HMAC:          &crypto.HMAC{},
	},
		udp.ServerWithClientRegisterHandler(func(u uuid.UUID) { fmt.Printf("\nuser %s registerd", u) }),
		udp.ServerWithReadBufferSize(2048),
		udp.ServerWithLogger(log.New(os.Stdout, "\n@Server Socket@------@", 1)),
		udp.ServerWithHeartbeatExpiration(time.Second),
		udp.ServerWithClientRequestHandler(func(u uuid.UUID, b1 byte, b2 []byte) { gameServer.ActionChan() <- []byte{1} }),
	)

	aesKey := []byte{113, 110, 25, 53, 11, 53, 68, 33, 17, 36, 22, 7, 125, 11, 35, 16, 83, 61, 59, 49, 31, 22, 69, 17, 24, 125, 11, 35, 16, 83, 61, 59}
	p1 := player1.GetID()
	client, _ := udp.NewClientServerManager(
		udp.ClientConfig{
			ServerAddr:         serverAddr,
			Encoder:            &udppb.Protobuf{},
			AsymmCrypto:        crypto.NewRSA(asymm),
			ServerAsymmPubKey:  rsaEnc.GetPublicKey(),
			SymmCrypto:         crypto.NewAESCBC(),
			ClientSymmKey:      aesKey,
			AuthToken:          p1[:],
			OnConnectionSucces: func() {},
			OnServerResponse: func(t byte, message []byte) {
				// m, _ := game_encoder.UnmarshalGameState(message)
				// fmt.Printf("\n#Client Socket One#------#server responeded with: %v, %v", t, m)
			},
			OnPingResult: func(i int64) {},
		},
		udp.ClientWithLogger(log.New(os.Stdout, "\n#Client Socket One#------#", 1)),
		udp.ClientWithPingInterval(500*time.Millisecond),
	)
	p2 := player2.GetID()
	client2, _ := udp.NewClientServerManager(
		udp.ClientConfig{
			ServerAddr:         serverAddr,
			Encoder:            &udppb.Protobuf{},
			AsymmCrypto:        crypto.NewRSA(asymm),
			ServerAsymmPubKey:  rsaEnc.GetPublicKey(),
			SymmCrypto:         crypto.NewAESCBC(),
			ClientSymmKey:      aesKey,
			AuthToken:          p2[:],
			OnConnectionSucces: func() {},
			OnServerResponse: func(t byte, message []byte) {
				m, _ := game_encoder.UnmarshalGameState(message)
				fmt.Printf("\n#Client Socket Two#------#server responeded with: %v, %v", t, m)
			},
			OnPingResult: func(i int64) {},
		},
		udp.ClientWithLogger(log.New(os.Stdout, "\n#Client Socket Two#------#", 1)),
		udp.ClientWithPingInterval(500*time.Millisecond),
	)

	go server.Serve()
	go gameServer.Start(2 * time.Second)
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

	for {
		select {
		case val := <-gameServer.StateChan():
			server.BroadcastToClients(9, val)
		case val := <-gameServer.EndChan():
			server.BroadcastToClients(10, val)
			time.Sleep(time.Second)
			return
			// default:
			// gameServer.ActionChan <- []byte{2, 4}

		}
	}
}
