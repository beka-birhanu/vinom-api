package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/beka-birhanu/vinom-api/api"
	gameapi "github.com/beka-birhanu/vinom-api/api/game"
	api_i "github.com/beka-birhanu/vinom-api/api/i"
	"github.com/beka-birhanu/vinom-api/api/identity"
	"github.com/beka-birhanu/vinom-api/config"
	grpc_matchmaking "github.com/beka-birhanu/vinom-api/infrastruture/grpc/matchmaking"
	grpc_sessionmanager "github.com/beka-birhanu/vinom-api/infrastruture/grpc/sessionmanager"
	"github.com/beka-birhanu/vinom-api/infrastruture/repo"
	"github.com/beka-birhanu/vinom-api/infrastruture/token"
	"github.com/beka-birhanu/vinom-api/service"
	"github.com/beka-birhanu/vinom-api/service/i"
	general_i "github.com/beka-birhanu/vinom-common/interfaces/general"
	logger "github.com/beka-birhanu/vinom-common/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Global variables for dependencies
var (
	sessionManagerGrpcConn *grpc.ClientConn
	matchmakerGrpcConn     *grpc.ClientConn
	mongoClient            *mongo.Client
	gameSessionManager     i.GameSessionManager
	userRepo               i.UserRepo
	matchmaker             i.Matchmaker
	matchmakingController  api_i.Controller
	jwtTokenizer           i.Tokenizer
	authService            i.Authenticator
	authController         api_i.Controller
	router                 *api.Router
	appLogger              general_i.Logger
)

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

func initUserRepo(client *mongo.Client) {
	userRepo = repo.NewUserRepo(client, config.Envs.DBName, "users")
	appLogger.Info("User repository initialized")
}

func initGrpcConns() {
	var err error
	matchmakingAddr := fmt.Sprintf("%s:%d", config.Envs.MatchmakingHost, config.Envs.MatchmakingPort)
	matchmakerGrpcConn, err = grpc.NewClient(matchmakingAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		appLogger.Error(fmt.Sprintf("Creating matchmaing gRPC connection : %v", err))
		os.Exit(1)
	}

	appLogger.Info("Created matchmaing gRPC connection")

	sessionmanagerAddr := fmt.Sprintf("%s:%d", config.Envs.SessionManagerHost, config.Envs.SessionManagerPort)
	sessionManagerGrpcConn, err = grpc.NewClient(sessionmanagerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		appLogger.Error(fmt.Sprintf("Creating session manager gRPC connection : %v", err))
		os.Exit(1)
	}

	appLogger.Info("Created session manager gRPC connection")
}

func initSessionManager() {
	sessionLogger, err := logger.New("SESSION-MANAGER", config.ColorCyan, os.Stdout)
	if err != nil {
		appLogger.Error(fmt.Sprintf("Creating session manager logger: %v", err))
		os.Exit(1)
	}

	gameSessionManager, err = grpc_sessionmanager.NewClient(sessionManagerGrpcConn, sessionLogger)
	if err != nil {
		appLogger.Error(fmt.Sprintf("Creating grpc session client: %v", err))
		os.Exit(1)
	}

	appLogger.Info("Session manager initialized")
}

func initMatchmaker() {
	matchLogger, err := logger.New("MATCH-MAKER", config.ColorPurple, os.Stdout)
	if err != nil {
		appLogger.Error(fmt.Sprintf("Creating matchmaker logger: %v", err))
		os.Exit(1)
	}

	matchmaker, err = grpc_matchmaking.NewClient(matchmakerGrpcConn, matchLogger)
	if err != nil {
		appLogger.Error(fmt.Sprintf("Creating grpc matchmaker client: %v", err))
		os.Exit(1)
	}

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
	jwtTokenizer = token.NewJwtService(config.Envs.JWTSecret, config.Envs.JWTIssuer)
	appLogger.Info("JWT Tokenizer initialized")
}

func initAuthService() {
	var err error
	authService, err = service.NewAuthService(userRepo, jwtTokenizer)
	if err != nil {
		appLogger.Error(fmt.Sprintf("Creating auth service: %v", err))
		os.Exit(1)
	}
	appLogger.Info("Auth service initialized")
}

func initAuthController() {
	authController = identity.NewIdentityServer(authService)
	appLogger.Info("Auth controller initialized")
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
	defer cancel() // Ensure the context is always canceled

	// Initialize dependencies
	appLogger, _ = logger.New("APP", config.ColorGreen, os.Stdout)

	initMongo(ctx)
	defer func() {
		_ = mongoClient.Disconnect(ctx)
	}()

	initUserRepo(mongoClient)
	initGrpcConns()
	defer sessionManagerGrpcConn.Close()
	defer matchmakerGrpcConn.Close()

	initSessionManager()
	initMatchmaker()
	initMatchmakingController()
	initJWTTokenizer()
	initAuthService()
	initAuthController()
	initRouter(jwtTokenizer)

	// Run HTTP server
	if err := router.Run(); err != nil {
		appLogger.Error(fmt.Sprintf("Starting server: %v", err))
		os.Exit(1)
	}

	// Allow time for cleanup operations (TODO: use WaitGroups instead)
	time.Sleep(2 * time.Second)
}
