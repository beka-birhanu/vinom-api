package maze

// Cell represents a single cell in a maze grid.
// It includes properties for walls on each side and an associated reward.
type Cell struct {
	// NorthWall indicates whether there is a wall on the north side of the cell.
	NorthWall bool
	// SouthWall indicates whether there is a wall on the south side of the cell.
	SouthWall bool
	// EastWall indicates whether there is a wall on the east side of the cell.
	EastWall bool
	// WestWall indicates whether there is a wall on the west side of the cell.
	WestWall bool
	// Reward specifies the reward value assigned to the cell.
	Reward int
}
