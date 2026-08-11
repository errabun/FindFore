package golfcourseapi

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseExternalID(t *testing.T) {
	assert.Equal(t, "42", parseExternalID(json.RawMessage(`42`)))
	assert.Equal(t, "abc", parseExternalID(json.RawMessage(`"abc"`)))
	assert.Equal(t, "", parseExternalID(json.RawMessage(`null`)))
	assert.Equal(t, "", parseExternalID(nil))
}
