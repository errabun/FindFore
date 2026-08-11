package entity

import "errors"

var (
	// ErrAlreadyOnEvent is returned when a player already has a player_events row for the event.
	ErrAlreadyOnEvent = errors.New("player is already part of this event")
	// ErrEventFull is returned when accepted count has reached open_spots (capacity).
	ErrEventFull = errors.New("event is full")
	// ErrEventMissing is returned when locking/joining a nonexistent event.
	ErrEventMissing = errors.New("event not found")
	// ErrPlayerEventMissing is returned when no player_events row exists for the pair.
	ErrPlayerEventMissing = errors.New("player event not found")
)
