package engine

type Cell struct {
	Character rune
	Color     uint8
}

var BlankCell = Cell{
	Character: ' ',
	Color:     0,
}
