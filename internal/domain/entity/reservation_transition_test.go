package entity_test

import (
	"testing"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/stretchr/testify/require"
)

func TestCanTransitionReservation(t *testing.T) {
	allowed := [][2]string{
		{entity.ReservationStatusPending, entity.ReservationStatusHeld},
		{entity.ReservationStatusPending, entity.ReservationStatusConfirmed},
		{entity.ReservationStatusPending, entity.ReservationStatusFailed},
		{entity.ReservationStatusHeld, entity.ReservationStatusConfirmed},
		{entity.ReservationStatusHeld, entity.ReservationStatusFailed},
		{entity.ReservationStatusHeld, entity.ReservationStatusCancelled},
		{entity.ReservationStatusConfirmed, entity.ReservationStatusCancelled},
	}
	for _, edge := range allowed {
		require.Truef(t, entity.CanTransitionReservation(edge[0], edge[1]), "%s -> %s", edge[0], edge[1])
	}

	disallowed := [][2]string{
		{entity.ReservationStatusFailed, entity.ReservationStatusConfirmed},
		{entity.ReservationStatusCancelled, entity.ReservationStatusConfirmed},
		{entity.ReservationStatusConfirmed, entity.ReservationStatusPending},
		{entity.ReservationStatusConfirmed, entity.ReservationStatusHeld},
		{entity.ReservationStatusPending, entity.ReservationStatusCancelled},
		{entity.ReservationStatusFailed, entity.ReservationStatusPending},
		{entity.ReservationStatusHeld, entity.ReservationStatusPending},
		{entity.ReservationStatusPending, entity.ReservationStatusPending},
	}
	for _, edge := range disallowed {
		require.Falsef(t, entity.CanTransitionReservation(edge[0], edge[1]), "%s -> %s", edge[0], edge[1])
	}
}

func TestTransitionReservation(t *testing.T) {
	res := &entity.Reservation{Status: entity.ReservationStatusPending}
	require.NoError(t, entity.TransitionReservation(res, entity.ReservationStatusHeld))
	require.Equal(t, entity.ReservationStatusHeld, res.Status)

	err := entity.TransitionReservation(res, entity.ReservationStatusPending)
	require.ErrorIs(t, err, entity.ErrInvalidReservationTransition)
	require.Equal(t, entity.ReservationStatusHeld, res.Status)
}
