package gamepb

import (
	"github.com/beka-birhanu/vinom-api/service/i"
	"google.golang.org/protobuf/proto"
)

var _ i.GameEncoder = &Protobuf{}

type Protobuf struct{}

// MarshalAction implements game.Encoder.
func (p *Protobuf) MarshalAction(a i.Action) ([]byte, error) {
	action := actionFromInterface(a)
	return proto.Marshal(action)
}

// MarshalCell implements game.Encoder.
func (p *Protobuf) MarshalCell(c i.Cell) ([]byte, error) {
	cell := cellFromInterface(c)
	return proto.Marshal(cell)
}

// MarshalCellPosition implements game.Encoder.
func (p *Protobuf) MarshalCellPosition(cp i.CellPosition) ([]byte, error) {
	cellPosition := cellPositionInterface(cp)
	return proto.Marshal(cellPosition)
}

// MarshalGameState implements game.Encoder.
func (p *Protobuf) MarshalGameState(gs i.GameState) ([]byte, error) {
	gameState := gameStateFromInterface(gs)
	return proto.Marshal(gameState)
}

// MarshalMaze implements game.Encoder.
func (p *Protobuf) MarshalMaze(m i.Maze) ([]byte, error) {
	maze := mazeFromInterface(m)
	return proto.Marshal(maze)
}

// MarshalPlayer implements game.Encoder.
func (p *Protobuf) MarshalPlayer(pl i.Player) ([]byte, error) {
	player := playerFromInterface(pl)
	return proto.Marshal(player)
}

// NewAction implements game.Encoder.
func (p *Protobuf) NewAction() i.Action {
	return &Action{}
}

// NewCell implements game.Encoder.
func (p *Protobuf) NewCell() i.Cell {
	return &Cell{}
}

// NewCellPosition implements game.Encoder.
func (p *Protobuf) NewCellPosition() i.CellPosition {
	return &Pos{}
}

// NewGameState implements game.Encoder.
func (p *Protobuf) NewGameState() i.GameState {
	return &GameState{}
}

// NewMaze implements game.Encoder.
func (p *Protobuf) NewMaze() i.Maze {
	return &Maze{}
}

// NewPlayer implements game.Encoder.
func (p *Protobuf) NewPlayer() i.Player {
	return &Player{}
}

// UnmarshalAction implements game.Encoder.
func (p *Protobuf) UnmarshalAction(b []byte) (i.Action, error) {
	action := &Action{}
	err := proto.Unmarshal(b, action)
	return action, err
}

// UnmarshalCell implements game.Encoder.
func (p *Protobuf) UnmarshalCell(b []byte) (i.Cell, error) {
	cell := &Cell{}
	err := proto.Unmarshal(b, cell)
	return cell, err
}

// UnmarshalCellPosition implements game.Encoder.
func (p *Protobuf) UnmarshalCellPosition(b []byte) (i.CellPosition, error) {
	pos := &Pos{}
	err := proto.Unmarshal(b, pos)
	return pos, err
}

// UnmarshalGameState implements game.Encoder.
func (p *Protobuf) UnmarshalGameState(b []byte) (i.GameState, error) {
	gameState := &GameState{}
	err := proto.Unmarshal(b, gameState)
	return gameState, err
}

// UnmarshalMaze implements game.Encoder.
func (p *Protobuf) UnmarshalMaze(b []byte) (i.Maze, error) {
	maze := &Maze{}
	err := proto.Unmarshal(b, maze)
	return maze, err
}

// UnmarshalPlayer implements game.Encoder.
func (p *Protobuf) UnmarshalPlayer(b []byte) (i.Player, error) {
	player := &Player{}
	err := proto.Unmarshal(b, player)
	return player, err
}
