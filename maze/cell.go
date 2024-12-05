package maze

type Cell struct {
	NorthWall bool
	SouthWall bool
	EastWall  bool
	WestWall  bool
	Reward    int
}
