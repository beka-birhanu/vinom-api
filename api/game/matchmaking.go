// Package gameapi handles game matchmaking and session management.
package gameapi

import (
	"context"
	"net/http"
	"time"

	"github.com/beka-birhanu/vinom-api/service/i"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MatchMakingController manages matchmaking operations.
type MatchMakingController struct {
	gameSessionManager i.GameSessionManager
	userRepo           i.UserRepo
	matchingService    i.Matchmaker
}

// NewMatchMakingController initializes a MatchMakingController.
func NewMatchMakingController(gsm i.GameSessionManager, ur i.UserRepo, ms i.Matchmaker) (*MatchMakingController, error) {
	return &MatchMakingController{
		gameSessionManager: gsm,
		userRepo:           ur,
		matchingService:    ms,
	}, nil
}

// RegisterPublic registers public routes.
func (mkc *MatchMakingController) RegisterPublic(route *gin.RouterGroup) {}

// RegisterProtected registers protected routes.
func (mkc *MatchMakingController) RegisterProtected(route *gin.RouterGroup) {
	matchMaking := route.Group("/gameMatch")
	{
		matchMaking.POST("/", mkc.match)
		matchMaking.GET("/:ID", mkc.matchInfo)
	}
}

// match handles match creation requests.
func (mkc *MatchMakingController) match(ctx *gin.Context) {
	//TODO: match id in ctx with request
	var request MatchRequest
	if err := ctx.ShouldBind(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	latency := time.Now().UnixMilli() - request.SentAt

	user, err := mkc.userRepo.ByID(request.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = mkc.matchingService.PushToQueue(context.Background(), user.ID, user.Rating, uint(latency))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error while matching player"})
		return
	}

	ctx.Status(http.StatusAccepted)
}

// matchInfo retrieves information about a specific match.
func (mkc *MatchMakingController) matchInfo(ctx *gin.Context) {
	//TODO: match id in ctx with request
	IDString := ctx.Params.ByName("ID")
	ID, err := uuid.Parse(IDString)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "id not found"})
		return
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	pubKey, socketAddr, err := mkc.gameSessionManager.SessionInfo(timeoutCtx, ID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "No Session"})
		return
	}

	response := &MatchInfoResponse{
		SocketPubKey: pubKey,
		SocketAddr:   socketAddr,
	}

	ctx.JSON(http.StatusOK, response)
}
