package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/beka-birhanu/vinom-api/api"
	gameapi "github.com/beka-birhanu/vinom-api/api/game"
	apii "github.com/beka-birhanu/vinom-api/api/i"
	"github.com/beka-birhanu/vinom-api/api/identity"
	"github.com/beka-birhanu/vinom-api/config"
	"github.com/beka-birhanu/vinom-api/infrastruture/crypto"
	gamepb "github.com/beka-birhanu/vinom-api/infrastruture/pb_encoder/game"
	udppb "github.com/beka-birhanu/vinom-api/infrastruture/pb_encoder/udp"
	"github.com/beka-birhanu/vinom-api/infrastruture/repo"
	"github.com/beka-birhanu/vinom-api/infrastruture/sortedstorage"
	"github.com/beka-birhanu/vinom-api/infrastruture/token"
	"github.com/beka-birhanu/vinom-api/infrastruture/udp"
	maze "github.com/beka-birhanu/vinom-api/infrastruture/willson_maze"
	"github.com/beka-birhanu/vinom-api/service"
	"github.com/beka-birhanu/vinom-api/service/i"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Global variables for dependencies
var (
	redisClient           *redis.Client
	mongoClient           *mongo.Client
	udpSocketManager      i.ServerSocketManager
	gameSessionManager    i.GameSessionManager
	userRepo              i.UserRepo
	matchmaker            i.Matchmaker
	matchmakingController apii.Controller
	jwtToken              i.Tokenizer
	authService           i.Authenticator
	authController        apii.Controller
	router                *api.Router
)

// Initialization functions
func initRedis(ctx context.Context) {
	addr := fmt.Sprintf("%s:%v", config.Envs.RedisHost, config.Envs.RedisPort)

	redisClient = redis.NewClient(&redis.Options{Addr: addr, DB: 0})
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		log.Fatalf("%s[APP] [ERROR] Failed to connect to Redis: %v%s", config.LogErrorColor, err, config.LogColorReset)
	}
	log.Printf("%s[APP] [INFO] Connected to Redis!%s", config.LogInfoColor, config.LogColorReset)
}

func initMongo(ctx context.Context) {
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%v", config.Envs.DBUser, config.Envs.DBPassword, config.Envs.DBHost, config.Envs.DBPort)

	clientOptions := options.Client().ApplyURI(uri)
	var err error
	mongoClient, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("%s[APP] [ERROR] Failed to connect to MongoDB: %v%s", config.LogErrorColor, err, config.LogColorReset)
	}
	if err = mongoClient.Ping(ctx, nil); err != nil {
		log.Fatalf("%s[APP] [ERROR] MongoDB ping failed: %v%s", config.LogErrorColor, uri, config.LogColorReset)
	}
	log.Printf("%s[APP] [INFO] Connected to MongoDB!%s", config.LogInfoColor, config.LogColorReset)
}

func initUDPSocketManager() {
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%v", config.Envs.HostIP, config.Envs.UDPPort))
	if err != nil {
		log.Fatalf("%s[APP] [ERROR] Resolving server address: %v%s", config.LogErrorColor, err, config.LogColorReset)
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("%s[APP] [ERROR] Generating RSA key: %v%s", config.LogErrorColor, err, config.LogColorReset)
	}

	rsaEnc := crypto.NewRSA(privateKey)
	server, err := udp.NewServerSocketManager(
		udp.ServerConfig{
			ListenAddr:  serverAddr,
			AsymmCrypto: rsaEnc,
			SymmCrypto:  crypto.NewAESCBC(),
			Encoder:     &udppb.Protobuf{},
			HMAC:        &crypto.HMAC{},
		},
		udp.ServerWithReadBufferSize(config.Envs.UDPBufferSize),
		udp.ServerWithLogger(log.New(os.Stdout, fmt.Sprintf("%s[SERVER-SOCKET] %s", config.ColorBlue, config.LogColorReset), log.LstdFlags|log.Lmsgprefix)),
		udp.ServerWithHeartbeatExpiration(3*time.Second),
	)
	if err != nil {
		log.Fatalf("%s[APP] [ERROR] Creating server UDP socket manager: %v%s", config.LogErrorColor, err, config.LogColorReset)
	}

	udpSocketManager = server
	log.Printf("%s[APP] [INFO] UDP Socket Manager initialized!%s", config.LogInfoColor, config.LogColorReset)
}

func initGameSessionManager(socketManager i.ServerSocketManager) {
	manager, err := service.NewGameSessionManager(
		&service.Config{
			Socket:       socketManager,
			MazeFactory:  maze.New,
			GameEndcoder: &gamepb.Protobuf{},
			Logger:       log.New(os.Stdout, fmt.Sprintf("%s[GAME-MANAGER] %s", config.ColorCyan, config.LogColorReset), log.LstdFlags|log.Lmsgprefix),
		},
	)
	if err != nil {
		log.Fatalf("%s[APP] [ERROR] Creating game session manager: %v%s", config.LogErrorColor, err, config.LogColorReset)
	}
	gameSessionManager = manager
	log.Printf("%s[APP] [INFO] Game Session Manager initialized!%s", config.LogInfoColor, config.LogColorReset)
}

func initUserRepo(client *mongo.Client) {
	userRepo = repo.NewUserRepo(client, config.Envs.DBName, "users")
	log.Printf("%s[APP] [INFO] User repository initialized!%s", config.LogInfoColor, config.LogColorReset)
}

func initMatchmaker(redisClient *redis.Client) {
	sortedQueue, err := sortedstorage.NewRedisSortedQueue(redisClient)
	if err != nil {
		log.Fatalf("%s[APP] [ERROR] Creating Redis sorted queue: %v%s", config.LogErrorColor, err, config.LogColorReset)
	}

	options := &service.Options{
		Logger:           log.New(os.Stdout, fmt.Sprintf("%s[MATCH-MAKER] %s", config.ColorMagenta, config.LogColorReset), log.LstdFlags|log.Lmsgprefix),
		MaxPlayer:        int64(config.Envs.MaxPlayer),
		RankTolerance:    config.Envs.RankTolerance,
		LatencyTolerance: config.Envs.LatencyTolerance,
	}

	maker, err := service.NewMatchmaker(sortedQueue, options)
	if err != nil {
		log.Fatalf("%s[APP] [ERROR] Creating matchmaker: %v%s", config.LogErrorColor, err, config.LogColorReset)
	}
	matchmaker = maker
	log.Printf("%s[APP] [INFO] Matchmaker initialized!%s", config.LogInfoColor, config.LogColorReset)
}

func initMatchmakingController() {
	var err error
	matchmakingController, err = gameapi.NewMatchMakingController(gameSessionManager, userRepo, matchmaker)
	if err != nil {
		log.Fatalf("%s[APP] [ERROR] Creating matchmaking controller: %v%s", config.LogErrorColor, err, config.LogColorReset)
	}
	log.Printf("%s[APP] [INFO] Matchmaking controller initialized!%s", config.LogInfoColor, config.LogColorReset)
}

func initJWTTokenizer() {
	jwtToken = token.NewJwtService(config.Envs.JWTSecret, config.Envs.JWTIssuer)
	log.Printf("%s[APP] [INFO] JWT Tokenizer initialized!%s", config.LogInfoColor, config.LogColorReset)
}

func initAuthService() {
	var err error
	authService, err = service.NewAuthService(userRepo, jwtToken)
	if err != nil {
		log.Fatalf("%s[APP] [ERROR] Creating auth service: %v%s", config.LogErrorColor, err, config.LogColorReset)
	}
	log.Printf("%s[APP] [INFO] Auth service initialized!%s", config.LogInfoColor, config.LogColorReset)
}

func initAuthController() {
	authController = identity.NewIdentityServer(authService)
	log.Printf("%s[APP] [INFO] Auth controller initialized!%s", config.LogInfoColor, config.LogColorReset)
}

func initRouter(t i.Tokenizer) {
	router = api.NewRouter(api.Config{
		Addr:                    fmt.Sprintf("%s:%v", config.Envs.HostIP, config.Envs.RESTPort),
		BaseURL:                 "/api",
		Controllers:             []apii.Controller{authController, matchmakingController},
		AuthorizationMiddleware: identity.Authoriz(t),
	})
	log.Printf("%s[APP] [INFO] Router initialized!%s", config.LogInfoColor, config.LogColorReset)
}

// TODO: add socket monitoring.
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Initialize dependencies
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
	log.Printf("%s[APP] [INFO] UDP Socket Manager started serving!%s", config.LogInfoColor, config.LogColorReset)

	if err := router.Run(); err != nil {
		log.Fatalf("%s[APP] [ERROR] Starting server: %v%s", config.LogErrorColor, err, config.LogColorReset)
	}
}
