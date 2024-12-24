package common

type Coordinate struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func (c Coordinate) Equals(other Coordinate) bool {
	return c.X == other.X && c.Y == other.Y
}
