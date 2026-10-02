package domain

import "time"

type GameFilter struct {
	Source   string
	Season   int
	Phase    string
	DateFrom *time.Time
	DateTo   *time.Time
	TeamID   int64
	Status   GameStatus
	Limit    int
	Offset   int
}
