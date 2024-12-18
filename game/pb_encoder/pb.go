package pb

import (
	"github.com/beka-birhanu/vinom-api/game"
	"google.golang.org/protobuf/proto"
)

var _ game.Encoder = &Protobuf{}

type Protobuf struct{}

// MarshalAction implements game.Encoder.
func (p *Protobuf) MarshalAction(a game.Action) ([]byte, error) {
	action := actionFromInterface(a)
	return proto.Marshal(action)
}

// MarshalCell implements game.Encoder.
func (p *Protobuf) MarshalCell(c game.Cell) ([]byte, error) {
	cell := cellFromInterface(c)
	return proto.Marshal(cell)
}

// MarshalCellPosition implements game.Encoder.
func (p *Protobuf) MarshalCellPosition(cp game.CellPosition) ([]byte, error) {
	cellPosition := cellPositionInterface(cp)
	return proto.Marshal(cellPosition)
}

// MarshalGameState implements game.Encoder.
func (p *Protobuf) MarshalGameState(gs game.GameState) ([]byte, error) {
	gameState := gameStateFromInterface(gs)
	return proto.Marshal(gameState)
}

// MarshalMaze implements game.Encoder.
func (p *Protobuf) MarshalMaze(m game.Maze) ([]byte, error) {
	maze := mazeFromInterface(m)
	return proto.Marshal(maze)
}

// MarshalPlayer implements game.Encoder.
func (p *Protobuf) MarshalPlayer(pl game.Player) ([]byte, error) {
	player := playerFromInterface(pl)
	return proto.Marshal(player)
}

// NewAction implements game.Encoder.
func (p *Protobuf) NewAction() game.Action {
	return &Action{}
}

// NewCell implements game.Encoder.
func (p *Protobuf) NewCell() game.Cell {
	return &Cell{}
}

// NewCellPosition implements game.Encoder.
func (p *Protobuf) NewCellPosition() game.CellPosition {
	return &Pos{}
}

// NewGameState implements game.Encoder.
func (p *Protobuf) NewGameState() game.GameState {
	return &GameState{}
}

// NewMaze implements game.Encoder.
func (p *Protobuf) NewMaze() game.Maze {
	return &Maze{}
}

// NewPlayer implements game.Encoder.
func (p *Protobuf) NewPlayer() game.Player {
	return &Player{}
}

// UnmarshalAction implements game.Encoder.
func (p *Protobuf) UnmarshalAction(b []byte) (game.Action, error) {
	action := &Action{}
	err := proto.Unmarshal(b, action)
	return action, err
}

// UnmarshalCell implements game.Encoder.
func (p *Protobuf) UnmarshalCell(b []byte) (game.Cell, error) {
	cell := &Cell{}
	err := proto.Unmarshal(b, cell)
	return cell, err
}

// UnmarshalCellPosition implements game.Encoder.
func (p *Protobuf) UnmarshalCellPosition(b []byte) (game.CellPosition, error) {
	pos := &Pos{}
	err := proto.Unmarshal(b, pos)
	return pos, err
}

// UnmarshalGameState implements game.Encoder.
func (p *Protobuf) UnmarshalGameState(b []byte) (game.GameState, error) {
	gameState := &GameState{}
	err := proto.Unmarshal(b, gameState)
	return gameState, err
}

// UnmarshalMaze implements game.Encoder.
func (p *Protobuf) UnmarshalMaze(b []byte) (game.Maze, error) {
	maze := &Maze{}
	err := proto.Unmarshal(b, maze)
	return maze, err
}

// UnmarshalPlayer implements game.Encoder.
func (p *Protobuf) UnmarshalPlayer(b []byte) (game.Player, error) {
	player := &Player{}
	err := proto.Unmarshal(b, player)
	return player, err
}

