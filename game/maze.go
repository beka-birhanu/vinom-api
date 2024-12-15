package game

// Cell defines the methods that a cell must implement.
type Cell interface {
	HasNorthWall() bool
	HasSouthWall() bool
	HasEastWall() bool
	HasWestWall() bool
	GetReward() int
	SetNorthWall(bool)
	SetSouthWall(bool)
	SetEastWall(bool)
	SetWestWall(bool)
	SetReward(int)
}

// CellPosition defines the methods that a cell position must implement.
type CellPosition interface {
	GetRow() int
	GetCol() int
	SetRow(int)
	SetCol(int)
}

// Move defines the methods for a move operation.
type Move interface {
	GetFrom() CellPosition
	GetTo() CellPosition
	GetDirection() string
	SetFrom(CellPosition)
	SetTo(CellPosition)
	SetDirection(string)
}

// Maze defines the methods that a maze must implement.
type Maze interface {
	IsValidMove(move Move) bool
	Move(move Move) (int, error)
	String() string
}
