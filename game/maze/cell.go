package maze

// CellInterface defines the methods that a cell must implement.
type CellInterface interface {
	// HasNorthWall returns true if there is a wall on the north side of the cell.
	HasNorthWall() bool

	// HasSouthWall returns true if there is a wall on the south side of the cell.
	HasSouthWall() bool

	// HasEastWall returns true if there is a wall on the east side of the cell.
	HasEastWall() bool

	// HasWestWall returns true if there is a wall on the west side of the cell.
	HasWestWall() bool

	// GetReward returns the reward value assigned to the cell.
	GetReward() int

	// SetNorthWall sets the presence of a wall on the north side of the cell.
	SetNorthWall(bool)

	// SetSouthWall sets the presence of a wall on the south side of the cell.
	SetSouthWall(bool)

	// SetEastWall sets the presence of a wall on the east side of the cell.
	SetEastWall(bool)

	// SetWestWall sets the presence of a wall on the west side of the cell.
	SetWestWall(bool)

	// SetReward sets the reward value assigned to the cell.
	SetReward(int)
}

// Cell represents a single cell in a maze grid.
// It includes properties for walls on each side and an associated reward.
type Cell struct {
	NorthWall bool // NorthWall indicates whether there is a wall on the north side of the cell.
	SouthWall bool // SouthWall indicates whether there is a wall on the south side of the cell.
	EastWall  bool // EastWall indicates whether there is a wall on the east side of the cell.
	WestWall  bool // WestWall indicates whether there is a wall on the west side of the cell.
	Reward    int  // Reward specifies the reward value assigned to the cell.
}

// HasNorthWall returns true if there is a wall on the north side of the cell.
func (c *Cell) HasNorthWall() bool {
	return c.NorthWall
}

// HasSouthWall returns true if there is a wall on the south side of the cell.
func (c *Cell) HasSouthWall() bool {
	return c.SouthWall
}

// HasEastWall returns true if there is a wall on the east side of the cell.
func (c *Cell) HasEastWall() bool {
	return c.EastWall
}

// HasWestWall returns true if there is a wall on the west side of the cell.
func (c *Cell) HasWestWall() bool {
	return c.WestWall
}

// GetReward returns the reward value assigned to the cell.
func (c *Cell) GetReward() int {
	return c.Reward
}

// SetNorthWall sets the presence of a wall on the north side of the cell.
func (c *Cell) SetNorthWall(hasWall bool) {
	c.NorthWall = hasWall
}

// SetSouthWall sets the presence of a wall on the south side of the cell.
func (c *Cell) SetSouthWall(hasWall bool) {
	c.SouthWall = hasWall
}

// SetEastWall sets the presence of a wall on the east side of the cell.
func (c *Cell) SetEastWall(hasWall bool) {
	c.EastWall = hasWall
}

// SetWestWall sets the presence of a wall on the west side of the cell.
func (c *Cell) SetWestWall(hasWall bool) {
	c.WestWall = hasWall
}

// SetReward sets the reward value assigned to the cell.
func (c *Cell) SetReward(reward int) {
	c.Reward = reward
}

// CellPosition represents the position of a cell in the maze grid.
type CellPosition struct {
	Row int // Row index of the cell
	Col int // Column index of the cell
}

// GetRow returns the row index of the cell.
func (cp *CellPosition) GetRow() int {
	return cp.Row
}

// GetCol returns the column index of the cell.
func (cp *CellPosition) GetCol() int {
	return cp.Col
}

// SetRow sets the row index of the cell.
func (cp *CellPosition) SetRow(row int) {
	cp.Row = row
}

// SetCol sets the column index of the cell.
func (cp *CellPosition) SetCol(col int) {
	cp.Col = col
}

// Move represents a movement from one cell to another in a specific direction.
type Move struct {
	From      CellPosition // Starting cell
	To        CellPosition // Destination cell
	Direction string       // Direction of the move (North, South, East, West)
}

// GetFrom returns the starting cell's position of the move.
func (m *Move) GetFrom() CellPosition {
	return m.From
}

// GetTo returns the destination cell's position of the move.
func (m *Move) GetTo() CellPosition {
	return m.To
}

// GetDirection returns the direction of the move (North, South, East, West).
func (m *Move) GetDirection() string {
	return m.Direction
}

// SetFrom sets the starting cell's position of the move.
func (m *Move) SetFrom(from CellPosition) {
	m.From = from
}

// SetTo sets the destination cell's position of the move.
func (m *Move) SetTo(to CellPosition) {
	m.To = to
}

// SetDirection sets the direction of the move (North, South, East, West).
func (m *Move) SetDirection(direction string) {
	m.Direction = direction
}

