package events_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ericrabun/findfore-go/internal/application/events"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func TestComposeAndSplitStartsAtRoundTrip(t *testing.T) {
	got, err := events.ComposeStartsAt("2026-08-15", "08:30", "America/Denver")
	require.NoError(t, err)

	date, teeTime, err := events.SplitStartsAt(got, "America/Denver")
	require.NoError(t, err)
	assert.Equal(t, "2026-08-15", date)
	assert.Equal(t, "08:30", teeTime)

	loc, err := time.LoadLocation("America/Denver")
	require.NoError(t, err)
	want := time.Date(2026, 8, 15, 8, 30, 0, 0, loc)
	assert.True(t, got.Equal(want))
	assert.Equal(t, want.UTC(), got.UTC())
}

func TestComposeStartsAtDefaultTimezone(t *testing.T) {
	got, err := events.ComposeStartsAt("2026-08-15", "09:30", "")
	require.NoError(t, err)

	loc, err := time.LoadLocation(entity.DefaultCourseTimezone)
	require.NoError(t, err)
	want := time.Date(2026, 8, 15, 9, 30, 0, 0, loc)
	assert.True(t, got.Equal(want))
}

func TestComposeStartsAtRejectsBadInput(t *testing.T) {
	_, err := events.ComposeStartsAt("", "08:00", "America/Denver")
	require.Error(t, err)
	_, err = events.ComposeStartsAt("2026-08-15", "", "America/Denver")
	require.Error(t, err)
	_, err = events.ComposeStartsAt("08-15-2026", "08:00", "America/Denver")
	require.Error(t, err)
}
