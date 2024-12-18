package game

import (
	"errors"
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Game-related errors.
var (
	ErrTooManyPlayers        = errors.New("too many players")
	ErrNotEnoughPlayers      = errors.New("not enough players")
	ErrNotBigEnoughDimension = errors.New("dimension is not big enough")
	ErrInvalidPlayerPosition = errors.New("player is out of the maze")
)

// Game constants for configuration and action types.
const (
	moveActionType         = 1 << iota // Action type for movement.
	stateRequestActionType             // Action type for state requests.

	minPlayers = 2 // Minimum number of players.
	maxPlayers = 4 // Maximum number of players.

	minDimension = 3 // Minimum maze dimension (width or height).
)

// Game represents a maze game with players, a maze, and game state.
// It manages player actions, broadcasts game state, and tracks game progress.
type Game struct {
	maze         Maze                 // The maze structure.
	players      map[uuid.UUID]Player // Map of players indexed by their IDs.
	version      int64                // Game state version for synchronization.
	encoder      Encoder              // Encoder for serializing game state.
	stop         chan bool            // stop channel to signal stop.
	rewardsLeft  int                  // Total rewards left in the maze.
	StateChan    chan []byte          // Channel for broadcasting state changes.
	ActionChan   chan []byte          // Channel for broadcasting actions.
	EndChan      chan []byte          // Channel to signal game completion.
	Wg           *sync.WaitGroup      // WaitGroup to manage concurrent goroutines.
	sync.RWMutex                      // Read-Write lock for synchronizing access.
}

// New creates a new Game instance with the specified maze, players, and encoder.
// Returns an error if configuration constraints are violated.
func New(maze Maze, players []Player, e Encoder) (*Game, error) {
	if len(players) > maxPlayers {
		return nil, ErrTooManyPlayers
	}

	if len(players) < minPlayers {
		return nil, ErrNotEnoughPlayers
	}

	if maze.Width() < minDimension || maze.Height() < minDimension {
		return nil, ErrNotBigEnoughDimension
	}

	playersMap := make(map[uuid.UUID]Player)
	for _, player := range players {
		if !maze.InBound(int(player.RetrivePos().GetRow()), int(player.RetrivePos().GetCol())) {
			return nil, ErrInvalidPlayerPosition
		}
		playersMap[player.GetID()] = player
		_ = maze.RemoveReward(player.RetrivePos())
	}

	return &Game{
		maze:        maze,
		players:     playersMap,
		rewardsLeft: maze.Width() * maze.Height(),
		encoder:     e,
		stop:        make(chan bool, 1),
		StateChan:   make(chan []byte),
		ActionChan:  make(chan []byte),
		EndChan:     make(chan []byte),
		Wg:          &sync.WaitGroup{},
	}, nil
}

// Start begins the game and listens for player actions or a timeout.
func (g *Game) Start(gameDuration time.Duration) {
	time.AfterFunc(gameDuration, g.Stop)
	for {
		select {
		case <-g.stop:
			close(g.stop)
			return
		case action := <-g.ActionChan:
			if len(action) < 2 {
				continue
			}
			g.handleAction(action[0], action[1:])
		}
	}
}

// handleAction processes incoming actions based on their type.
func (g *Game) handleAction(t byte, move []byte) {
	switch t {
	case stateRequestActionType:
		g.Wg.Add(1)
		go g.broadcastState(false)
	case moveActionType:
		a, err := g.encoder.UnmarshalAction(move)
		if err != nil {
			return
		}
		go g.handleIncomingMove(a)
	}
}

// Stop ends the game, closes channels, and broadcasts the final state.
func (g *Game) Stop() {
	g.stop <- true
	g.Wg.Wait()
	close(g.ActionChan)
	close(g.StateChan)
	g.Wg.Add(1)
	g.broadcastState(true)
	close(g.EndChan)
}

// broadcastState sends the current game state to all players.
func (g *Game) broadcastState(ended bool) {
	defer g.Wg.Done()
	gameState := g.snapshot()
	gameStatePayload, err := g.encoder.MarshalGameState(gameState)
	if err != nil {
		return
	}

	if ended {
		g.EndChan <- gameStatePayload
	} else {
		g.StateChan <- gameStatePayload
	}
}

// snapshot creates a snapshot of the current game state.
func (g *Game) snapshot() GameState {
	g.RLock()
	defer g.RUnlock()

	gameState := g.encoder.NewGameState()
	gameState.SetVersion(g.version)
	gameState.SetMaze(g.maze)
	gameState.SetPlayers(slices.Collect(maps.Values(g.players)))
	return gameState
}

// handleIncomingMove processes player movement actions.
// It validates the move, updates the state, and broadcasts changes.
func (g *Game) handleIncomingMove(a Action) {
	g.Lock()
	p, ok := g.players[a.GetID()]
	if !ok {
		return
	}

	curPosition := p.RetrivePos()
	if curPosition.GetRow() != a.RetriveFrom().GetRow() || curPosition.GetCol() != a.RetriveFrom().GetCol() {
		return
	}

	move, err := g.maze.NewValidMove(curPosition, a.GetDirection())
	if err != nil {
		g.Unlock()
		return
	}

	reward, _ := g.maze.Move(move)
	p.SetReward(p.GetReward() + reward)
	g.version++
	if g.maze.GetTotalReward() == 0 {
		g.Unlock()
		g.Stop()
		return
	}
	g.Unlock()

	g.Wg.Add(1)
	go g.broadcastState(false)
}
