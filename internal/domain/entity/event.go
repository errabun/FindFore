package entity

import "time"

type Event struct {
	ID              int64
	CourseID        int32
	Date            string // API wall-clock input/output (not a DB column)
	TeeTime         string // API wall-clock input/output (not a DB column)
	PlannedStartsAt time.Time
	TeeTimeID       *int64
	GroupID         *int64
	OpenSpots       int32
	NumberOfHoles   string
	Private         bool
	HostID          int32
}

type EventWithDetails struct {
	ID              int64
	CourseName      string
	CourseTimezone  string
	Date            string
	TeeTime         string
	PlannedStartsAt time.Time
	TeeTimeID       *int64
	GroupID         *int64
	OpenSpots       int32
	NumberOfHoles   string
	Private         bool
	HostName        string
	HostID          int32
	Accepted        []int64
	Declined        []int64
	Pending         []int64
	Closed          []int64
	RemainingSpots  int32
}
