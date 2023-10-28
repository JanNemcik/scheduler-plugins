package networktraffic

import (
	"strconv"
)

const (
	Lowest  = 0
	Low     = 1
	High    = 2
	Highest = 3
	Lowest2 = 10
	Low2    = 50
	High2   = 80
)

var ContextPlacementMap = map[int]int{10: Lowest, 50: Low, 80: High}
var ContextPlacementMap2 = map[string]int{"Lowest": Lowest2, "Low": Low2, "High": High2}

func getContext(speed string) {
	var transSpeed float64
	if s, err := strconv.ParseFloat(speed, 64); err == nil {
		transSpeed = s
	}
	switch transSpeed {
	case float64(Lowest):

	}
}
