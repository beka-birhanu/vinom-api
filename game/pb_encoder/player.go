package pb

import (
	"github.com/beka-birhanu/vinom-api/game"
	"github.com/google/uuid"
)

var _ game.Player = &Player{}

func playerFromInterface(player game.Player) *Player {
	return &Player{
		Pos:    cellPositionInterface(player.RetrivePos()),
		Reward: player.GetReward(),
		Id:     player.GetID().String(),
	}
}

// GetID implements game.Player.
func (x *Player) GetID() uuid.UUID {
	id, _ := uuid.FromBytes([]byte(x.Id))
	return id
}

// RetrivePos implements game.Player.
func (x *Player) RetrivePos() game.CellPosition {
	return x.Pos
}

// SetID implements game.Player.
func (x *Player) SetID(i uuid.UUID) {
	x.Id = i.String()
}

// SetPos implements game.Player.
func (x *Player) SetPos(p game.CellPosition) {
	x.Pos = cellPositionInterface(x.Pos)
}

// SetReward implements game.Player.
func (x *Player) SetReward(r int32) {
	x.Reward = r
}
