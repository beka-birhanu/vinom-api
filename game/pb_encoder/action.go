package pb

import (
	"github.com/beka-birhanu/vinom-api/game"
	"github.com/google/uuid"
)

var _ game.Action = &Action{}

func actionFromInterface(a game.Action) *Action {
	return &Action{
		Id:        a.GetID().String(),
		Direction: a.GetDirection(),
		From:      cellPositionInterface(a.RetriveFrom()),
	}
}

// RetriveFrom implements game.Action.
func (x *Action) RetriveFrom() game.CellPosition {
	return x.From
}

// SetFrom implements game.Action.
func (x *Action) SetFrom(c game.CellPosition) {
	x.From = cellPositionInterface(c)
}

// SetDirection implements game.Action.
func (x *Action) SetDirection(s string) {
	x.Direction = s
}

// GetID implements game.Action.
func (x *Action) GetID() uuid.UUID {
	id, _ := uuid.FromBytes([]byte(x.Id))
	return id
}

// SetID implements game.Action.
func (x *Action) SetID(i uuid.UUID) {
	x.Id = i.String()
}
