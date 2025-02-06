package service

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/beka-birhanu/vinom-api/config"
	"github.com/beka-birhanu/vinom-api/service/i"
	"github.com/google/uuid"
)

const (
	defaultMazeSize     = 10
	defaultGameDuration = 5 * time.Minute

	gameStateRecordType = 10
	gameEndedRecordType = 11
)

var (
	defaultPlayerPositions = []struct {
		row int32
		col int32
	}{{row: 0, col: 0}, {row: 9, col: 9}, {row: 9, col: 0}, {row: 0, col: 9}}
)

type GameSessionManager struct {
	socket   i.ServerSocketManager
	sessions map[uuid.UUID]struct {
		gameSession i.GameServer
		players     []uuid.UUID
	}
	playerToSession map[uuid.UUID]uuid.UUID
	mazeFactory     func(int, int) (i.Maze, error)
	gameEndcoder    i.GameEncoder
	logger          *log.Logger
	sync.RWMutex
}

type Config struct {
	Socket       i.ServerSocketManager
	MazeFactory  func(int, int) (i.Maze, error)
	GameEndcoder i.GameEncoder
	Logger       *log.Logger
}

func NewGameSessionManager(c *Config) (*GameSessionManager, error) {
	gsm := &GameSessionManager{
		socket:       c.Socket,
		gameEndcoder: c.GameEndcoder,
		mazeFactory:  c.MazeFactory,
		logger:       c.Logger,
		sessions: make(map[uuid.UUID]struct {
			gameSession i.GameServer
			players     []uuid.UUID
		}),
		playerToSession: make(map[uuid.UUID]uuid.UUID),
	}

	c.Socket.SetClientRequestHandler(gsm.writePlayerRequest)
	c.Socket.SetClientAuthenticator(gsm)
	return gsm, nil
}

func (g *GameSessionManager) NewSession(playerIDs []uuid.UUID) {
	if len(playerIDs) > maxPlayers {
		g.logger.Printf("%s[ERROR]%s too many players in game session: %d", config.LogErrorColor, config.LogColorReset, len(playerIDs))
		return
	}

	players := make([]i.Player, 0)
	for i, pID := range playerIDs {
		pos := g.gameEndcoder.NewCellPosition()
		pos.SetRow(defaultPlayerPositions[i].row)
		pos.SetCol(defaultPlayerPositions[i].col)

		player := g.gameEndcoder.NewPlayer()
		player.SetID(pID)
		player.SetPos(pos)
		players = append(players, player)
	}

	maze, err := g.mazeFactory(20, defaultMazeSize)
	if err != nil {
		g.logger.Printf("%s[ERROR]%s creating maze for a new game: %s", config.LogErrorColor, config.LogColorReset, err)
		return
	}

	mazeRewardModel := struct {
		RewardOne      int32
		RewardTwo      int32
		RewardTypeProb float32
	}{RewardOne: 1, RewardTwo: 5, RewardTypeProb: 0.9}

	if err := maze.PopulateReward(mazeRewardModel); err != nil {
		g.logger.Printf("%s[ERROR]%s populating rewards for a new game: %s", config.LogErrorColor, config.LogColorReset, err)
		return
	}

	gameServer, err := NewGame(maze, players, g.gameEndcoder)
	if err != nil {
		g.logger.Printf("%s[ERROR]%s creating new game server: %s", config.LogErrorColor, config.LogColorReset, err)
		return
	}

	sessionID := g.saveSession(playerIDs, gameServer)
	go gameServer.Start(defaultGameDuration)
	go g.listenGameChan(sessionID)
	g.logger.Printf("%s[INFO]%s started new game for players: %v", config.LogInfoColor, config.LogColorReset, playerIDs)
}

func (g *GameSessionManager) SessionInfo(playerID uuid.UUID) ([]byte, string, error) {
	g.RLock()
	defer g.RUnlock()
	if _, ok := g.playerToSession[playerID]; !ok {
		return nil, "", errors.New("No Session")
	}
	return g.socket.GetPublicKey(), g.socket.GetAddr(), nil
}

func (g *GameSessionManager) Authenticate(s []byte) (uuid.UUID, error) {
	g.RLock()
	defer g.RUnlock()
	id, err := uuid.FromBytes(s)
	if err != nil {
		g.logger.Printf("%s[ERROR]%s invalid token provided", config.LogErrorColor, config.LogColorReset)
		return uuid.Nil, errors.New("invalid token")
	}

	if _, ok := g.playerToSession[id]; !ok {
		g.logger.Printf("%s[ERROR]%s player does not have a game session", config.LogErrorColor, config.LogColorReset)
		return uuid.Nil, errors.New("player does not have game session")
	}

	g.logger.Printf("%s[INFO]%s authenticated player: %s", config.LogInfoColor, config.LogColorReset, id)
	return id, nil
}

func (g *GameSessionManager) saveSession(players []uuid.UUID, gs i.GameServer) uuid.UUID {
	sessionID := uuid.New()
	for {
		if _, ok := g.sessions[sessionID]; !ok {
			break
		}
		sessionID = uuid.New()
	}

	g.sessions[sessionID] = struct {
		gameSession i.GameServer
		players     []uuid.UUID
	}{gameSession: gs, players: players}
	for _, player := range players {
		g.playerToSession[player] = sessionID
	}

	return sessionID
}

func (g *GameSessionManager) listenGameChan(id uuid.UUID) {
	gs := g.sessions[id].gameSession
	players := g.sessions[id].players
	for {
		select {
		case val, ok := <-gs.StateChan():
			if !ok {
				break
			}
			g.socket.BroadcastToClients(players, gameStateRecordType, val)
		case val, ok := <-gs.EndChan():
			if !ok {
				break
			}
			g.socket.BroadcastToClients(players, gameEndedRecordType, val)
			g.clean(id)
			return
		}
	}
}

func (g *GameSessionManager) writePlayerRequest(pID uuid.UUID, actionType byte, payload []byte) {
	g.RLock()
	defer g.RUnlock()
	sessionID, ok := g.playerToSession[pID]
	if !ok {
		g.logger.Printf("%s[ERROR]%s received request for player without session", config.LogErrorColor, config.LogColorReset)
		return
	}

	gameServer := g.sessions[sessionID].gameSession
	gameServer.ActionChan() <- append([]byte{actionType}, payload...)
	g.logger.Printf("%s[INFO]%s processed request for player: %s", config.LogInfoColor, config.LogColorReset, pID)
}

func (g *GameSessionManager) clean(ID uuid.UUID) {
	g.Lock()
	defer g.Unlock()
	for _, pID := range g.sessions[ID].players {
		delete(g.playerToSession, pID)
	}

	delete(g.sessions, ID)
}

func (g *GameSessionManager) StopAll() {
	g.Lock()
	defer g.Unlock()

	for _, session := range g.sessions {
		session.gameSession.Stop()
	}
}
