package streamchat

import (
	"context"
	"fmt"

	"github.com/GetStream/getstream-go/v4"
)

func (a *Adapter) ensureMessagingFeatures(ctx context.Context) {
	a.featuresMu.Lock()
	defer a.featuresMu.Unlock()
	if a.featuresReady {
		return
	}
	if err := a.configureMessagingType(ctx); err != nil {
		return
	}
	a.featuresReady = true
}

func (a *Adapter) configureMessagingType(ctx context.Context) error {
	resp, err := a.client.Chat().GetChannelType(ctx, channelType, &getstream.GetChannelTypeRequest{})
	if err != nil {
		return fmt.Errorf("get messaging channel type: %w", err)
	}
	cfg := resp.Data
	commands := make([]string, 0, len(cfg.Commands)+1)
	hasGiphy := false
	for _, cmd := range cfg.Commands {
		if cmd.Name == "" {
			continue
		}
		commands = append(commands, cmd.Name)
		if cmd.Name == "giphy" {
			hasGiphy = true
		}
	}
	if !hasGiphy {
		commands = append(commands, "giphy")
	}
	if cfg.Uploads && hasGiphy {
		return nil
	}
	_, err = a.client.Chat().UpdateChannelType(ctx, channelType, &getstream.UpdateChannelTypeRequest{
		Automod:          cfg.Automod,
		AutomodBehavior:  cfg.AutomodBehavior,
		MaxMessageLength: cfg.MaxMessageLength,
		Uploads:          getstream.PtrTo(true),
		Commands:         commands,
		Permissions:      cfg.Permissions,
		Grants:           cfg.Grants,
	})
	if err != nil {
		return fmt.Errorf("update messaging channel type: %w", err)
	}
	return nil
}
