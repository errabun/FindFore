package httphandler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	httphandler "github.com/ericrabun/findfore-go/internal/adapter/inbound/http"
	"github.com/ericrabun/findfore-go/internal/application/groups"
	"github.com/ericrabun/findfore-go/internal/auth"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

type stubChat struct {
	sess *port.GroupChatSession
	err  error
}

func (s stubChat) GroupSession(context.Context, int64, int64) (*port.GroupChatSession, error) {
	return s.sess, s.err
}

func chatRouter(t *testing.T, chat port.ChatService) http.Handler {
	t.Helper()
	h := httphandler.New(stubPlayers{}, stubSessions{}, stubCourses{}, stubEvents{}, stubPlayerEvents{}, stubFriendships{}, stubPosts{}, nil, nil)
	if chat != nil {
		h = h.WithChat(chat)
	}
	return httphandler.NewRouter(h, testJWTSecret, stubTokenVersions{versions: map[int64]int32{1: 0, 2: 0}})
}

func chatRequest(t *testing.T, router http.Handler, path string, playerID int64) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, path, nil)
	if playerID > 0 {
		tok, err := auth.GenerateToken(playerID, 0, testJWTSecret)
		require.NoError(t, err)
		r.Header.Set("Authorization", "Bearer "+tok)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, r)
	return rec
}

func TestGroupChatHTTPUnauthorized(t *testing.T) {
	rec := chatRequest(t, chatRouter(t, stubChat{}), "/api/v1/groups/1/chat", 0)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestGroupChatHTTPUnconfigured(t *testing.T) {
	rec := chatRequest(t, chatRouter(t, nil), "/api/v1/groups/1/chat", 1)
	require.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

func TestGroupChatHTTPNotFoundForNonMember(t *testing.T) {
	rec := chatRequest(t, chatRouter(t, stubChat{err: groups.ErrGroupNotFound}), "/api/v1/groups/1/chat", 2)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGroupChatHTTPMemberSession(t *testing.T) {
	rec := chatRequest(t, chatRouter(t, stubChat{
		sess: &port.GroupChatSession{
			APIKey: "pk", Token: "tok", ChannelType: "messaging",
			ChannelID: "group_1", UserID: "1", UserName: "Eric",
		},
	}), "/api/v1/groups/1/chat", 1)
	require.Equal(t, http.StatusOK, rec.Code)
	var body groupChatJSON
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "pk", body.APIKey)
	require.Equal(t, "tok", body.Token)
	require.Equal(t, "messaging", body.ChannelType)
	require.Equal(t, "group_1", body.ChannelID)
	require.Equal(t, "1", body.UserID)
	require.Equal(t, "Eric", body.UserName)
}

type groupChatJSON struct {
	APIKey      string `json:"api_key"`
	Token       string `json:"token"`
	ChannelType string `json:"channel_type"`
	ChannelID   string `json:"channel_id"`
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
}
