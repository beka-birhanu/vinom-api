package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net"
	"os"
	"time"

	crypto "github.com/beka-birhanu/udp-socket-manager/crypto"
	udppb "github.com/beka-birhanu/udp-socket-manager/encoding"
	udpsocket "github.com/beka-birhanu/udp-socket-manager/socket"
	"github.com/beka-birhanu/vinom-api/api"
	gameapi "github.com/beka-birhanu/vinom-api/api/game"
	api_i "github.com/beka-birhanu/vinom-api/api/i"
	"github.com/beka-birhanu/vinom-api/api/identity"
	"github.com/beka-birhanu/vinom-api/config"
	logger "github.com/beka-birhanu/vinom-api/infrastruture/log"
	"github.com/beka-birhanu/vinom-api/infrastruture/repo"
	"github.com/beka-birhanu/vinom-api/infrastruture/sortedstorage"
	"github.com/beka-birhanu/vinom-api/infrastruture/token"
	"github.com/beka-birhanu/vinom-api/service"
	"github.com/beka-birhanu/vinom-api/service/i"
	gamepb "github.com/beka-birhanu/vinom-game-encoder"
	general_i "github.com/beka-birhanu/vinom-interfaces/general"
	socket_i "github.com/beka-birhanu/vinom-interfaces/socket"
	maze "github.com/beka-birhanu/wilson-maze"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Global variables for dependencies
var (
	redisClient           *redis.Client
	mongoClient           *mongo.Client
	udpSocketManager      socket_i.ServerSocketManager
	gameSessionManager    i.GameSessionManager
	userRepo              i.UserRepo
	matchmaker            i.Matchmaker
	matchmakingController api_i.Controller
	jwtToken              i.Tokenizer
	authService           i.Authenticator
	authController        api_i.Controller
	router                *api.Router
	appLogger             general_i.Logger
)

// Initialization functions
func initRedis(ctx context.Context) {
	addr := fmt.Sprintf("%s:%v", config.Envs.RedisHost, config.Envs.RedisPort)

	redisClient = redis.NewClient(&redis.Options{Addr: addr, DB: 0})
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		appLogger.Error(fmt.Sprintf("Failed to connect to Redis: %v", err))
		os.Exit(1)
	}
	appLogger.Info("Connected to Redis")
}

func initMongo(ctx context.Context) {
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%v", config.Envs.DBUser, config.Envs.DBPassword, config.Envs.DBHost, config.Envs.DBPort)

	clientOptions := options.Client().ApplyURI(uri)
	var err error
	mongoClient, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		appLogger.Error(fmt.Sprintf("Failed to connect to MongoDB: %v", err))
		os.Exit(1)
	}
	if err = mongoClient.Ping(ctx, nil); err != nil {
		appLogger.Error(fmt.Sprintf("MongoDB ping failed: %v", err))
		os.Exit(1)
	}
	appLogger.Info("Connected to MongoDB")
}

func initUDPSocketManager() {
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%v", config.Envs.HostIP, config.Envs.UDPPort))
	if err != nil {
		appLogger.Error(fmt.Sprintf("Resolving server address: %v", err))
		os.Exit(1)
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		appLogger.Error(fmt.Sprintf("Generating RSA key: %v", err))
		os.Exit(1)
	}

	rsaEnc := crypto.NewRSA(privateKey)

	serverLogger, err := logger.New("SERVER-SOCKET", config.ColorBlue, os.Stdout)
	if err != nil {
		appLogger.Error(fmt.Sprintf("Creating UDP socket manager logger: %v", err))
		os.Exit(1)
	}
	server, err := udpsocket.NewServerSocketManager(
		udpsocket.ServerConfig{
			ListenAddr:  serverAddr,
			AsymmCrypto: rsaEnc,
			SymmCrypto:  crypto.NewAESCBC(),
			Encoder:     &udppb.Protobuf{},
			HMAC:        &crypto.HMAC{},
			Logger:      serverLogger,
		},
		udpsocket.ServerWithReadBufferSize(config.Envs.UDPBufferSize),
		udpsocket.ServerWithHeartbeatExpiration(3*time.Second),
	)
	if err != nil {
		appLogger.Error(fmt.Sprintf("Creating server UDP socket manager: %v", err))
		os.Exit(1)
	}

	udpSocketManager = server
	appLogger.Info("UDP Socket Manager initialized")
}

func initGameSessionManager(socketManager socket_i.ServerSocketManager) {
	gameLogger, err := logger.New("GAME-MANAGER", config.ColorCyan, os.Stdout)
	if err != nil {
		appLogger.Error(fmt.Sprintf("Creating UDP socket manager logger: %v", err))
		os.Exit(1)
	}
	manager, err := service.NewGameSessionManager(
		&service.Config{
			Socket:       socketManager,
			MazeFactory:  maze.New,
			GameEndcoder: &gamepb.Protobuf{},
			Logger:       gameLogger,
		},
	)
	if err != nil {
		appLogger.Error(fmt.Sprintf("Creating game session manager: %v", err))
		os.Exit(1)
	}
	gameSessionManager = manager
	appLogger.Info("Game Session Manager initialized")
}

func initUserRepo(client *mongo.Client) {
	userRepo = repo.NewUserRepo(client, config.Envs.DBName, "users")
	appLogger.Info("User repository initialized")
}

func initMatchmaker(redisClient *redis.Client) {
	sortedQueue, err := sortedstorage.NewRedisSortedQueue(redisClient, 300)
	if err != nil {
		appLogger.Error(fmt.Sprintf("Creating Redis sorted queue: %v", err))
		os.Exit(1)
	}
	matchLogger, err := logger.New("MATCH-MAKER", config.ColorPurple, os.Stdout)
	if err != nil {
		appLogger.Error(fmt.Sprintf("Creating matchmaker logger: %v", err))
		os.Exit(1)
	}
	options := &service.Options{
		MaxPlayer:        int64(config.Envs.MaxPlayer),
		RankTolerance:    config.Envs.RankTolerance,
		LatencyTolerance: config.Envs.LatencyTolerance,
	}

	maker, err := service.NewMatchmaker(sortedQueue, matchLogger, options)
	if err != nil {
		appLogger.Error(fmt.Sprintf("Creating matchmaker: %v", err))
		os.Exit(1)
	}
	matchmaker = maker
	appLogger.Info("Matchmaker initialized")
}

func initMatchmakingController() {
	var err error
	matchmakingController, err = gameapi.NewMatchMakingController(gameSessionManager, userRepo, matchmaker)
	if err != nil {
		appLogger.Error(fmt.Sprintf("Creating matchmaking controller: %v", err))
		os.Exit(1)
	}
	appLogger.Info("Matchmaking controller initialized")
}

func initJWTTokenizer() {
	jwtToken = token.NewJwtService(config.Envs.JWTSecret, config.Envs.JWTIssuer)
	appLogger.Info("JWT Tokenizer initialized")
}

func initAuthService() {
	var err error
	authService, err = service.NewAuthService(userRepo, jwtToken)
	if err != nil {
		appLogger.Error(fmt.Sprintf("Creating auth service: %v", err))
		os.Exit(1)
	}
	appLogger.Info("Auth service initialized")
}

func initAuthController() {
	authController = identity.NewIdentityServer(authService)
	appLogger.Info("Auth controller initialized initialized")
}

func initRouter(t i.Tokenizer) {
	router = api.NewRouter(api.Config{
		Addr:                    fmt.Sprintf("%s:%v", config.Envs.HostIP, config.Envs.RESTPort),
		BaseURL:                 "/api",
		Controllers:             []api_i.Controller{authController, matchmakingController},
		AuthorizationMiddleware: identity.Authoriz(t),
	})
	appLogger.Info("Router initialized")
}

// TODO: add socket monitoring.
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Initialize dependencies
	appLogger, _ = logger.New("APP", config.ColorGreen, os.Stdout)
	initRedis(ctx)
	initMongo(ctx)
	initUDPSocketManager()
	initGameSessionManager(udpSocketManager)
	initUserRepo(mongoClient)
	initMatchmaker(redisClient)
	initMatchmakingController()
	initJWTTokenizer()
	initAuthService()
	initAuthController()
	initRouter(jwtToken)

	defer func() {
		_ = redisClient.Close()
		_ = mongoClient.Disconnect(ctx)
		gameSessionManager.StopAll()
		time.Sleep(2 * time.Second) // Wait for all games to finish sending data. @TODO: use better way to wait, maybe waitgroups.
		udpSocketManager.Stop()
	}()

	go udpSocketManager.Serve()
	appLogger.Info("UDP Socket Manager started serving")

	if err := router.Run(); err != nil {
		appLogger.Error(fmt.Sprintf("Starting server: %v", err))
		os.Exit(1)
	}
}
