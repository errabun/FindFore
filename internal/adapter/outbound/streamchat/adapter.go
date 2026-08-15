package streamchat

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/GetStream/getstream-go/v4"

	"github.com/ericrabun/findfore-go/internal/domain/port"
)

const channelType = "messaging"

type Adapter struct {
	client        *getstream.Stream
	apiKey        string
	featuresMu    sync.Mutex
	featuresReady bool
}

func New(apiKey, apiSecret string) (*Adapter, error) {
	if apiKey == "" || apiSecret == "" {
		return nil, fmt.Errorf("stream api key and secret are required")
	}
	client, err := getstream.NewClient(apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("stream client: %w", err)
	}
	return &Adapter{client: client, apiKey: apiKey}, nil
}

func (a *Adapter) APIKey() string { return a.apiKey }

func (a *Adapter) IssueToken(ctx context.Context, playerID int64, name string) (string, error) {
	userID := strconv.FormatInt(playerID, 10)
	users := map[string]getstream.UserRequest{
		userID: {ID: userID, Name: getstream.PtrTo(name)},
	}
	if _, err := a.client.UpdateUsers(ctx, &getstream.UpdateUsersRequest{Users: users}); err != nil {
		return "", fmt.Errorf("upsert stream user: %w", err)
	}
	token, err := a.client.CreateToken(userID, getstream.WithExpiration(tokenTTL))
	if err != nil {
		return "", fmt.Errorf("create stream token: %w", err)
	}
	return token, nil
}

func (a *Adapter) EnsureGroupChannel(ctx context.Context, groupID int64, name string, members []port.GroupMember) error {
	a.ensureMessagingFeatures(ctx)
	if len(members) == 0 {
		return fmt.Errorf("group %d has no members", groupID)
	}
	users := make(map[string]getstream.UserRequest, len(members))
	memberReqs := make([]getstream.ChannelMemberRequest, 0, len(members))
	for _, m := range members {
		id := strconv.FormatInt(m.PlayerID, 10)
		users[id] = getstream.UserRequest{ID: id, Name: getstream.PtrTo(m.PlayerName)}
		memberReqs = append(memberReqs, getstream.ChannelMemberRequest{UserID: id})
	}
	if _, err := a.client.UpdateUsers(ctx, &getstream.UpdateUsersRequest{Users: users}); err != nil {
		return fmt.Errorf("upsert stream members: %w", err)
	}

	createdBy := strconv.FormatInt(members[0].PlayerID, 10)
	ch := a.client.Chat().Channel(channelType, channelID(groupID))
	if _, err := ch.GetOrCreate(ctx, &getstream.GetOrCreateChannelRequest{
		Data: &getstream.ChannelInput{
			CreatedByID: getstream.PtrTo(createdBy),
			Members:     memberReqs,
			Custom:      map[string]any{"name": name},
		},
	}); err != nil {
		return fmt.Errorf("get or create stream channel: %w", err)
	}
	if _, err := ch.Update(ctx, &getstream.UpdateChannelRequest{
		AddMembers: memberReqs,
		Data: &getstream.ChannelInputRequest{
			Custom: map[string]any{"name": name},
		},
	}); err != nil {
		return fmt.Errorf("sync stream channel members: %w", err)
	}
	return nil
}

func channelID(groupID int64) string {
	return fmt.Sprintf("group_%d", groupID)
}
