package groups_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

func TestListDiscoverPaginationSearchAndOrder(t *testing.T) {
	svc, _ := newSvc()
	ctx := context.Background()
	names := []string{"Zebra Golf", "Alpha Golf", "Alpha Golf", "Mid Crew"}
	for _, name := range names {
		_, err := svc.Create(ctx, port.CreateGroupInput{
			ActorID: ownerID, Name: name, Privacy: entity.GroupPrivacyPublic,
		})
		require.NoError(t, err)
	}
	_, err := svc.Create(ctx, port.CreateGroupInput{
		ActorID: ownerID, Name: "Secret Crew", Privacy: entity.GroupPrivacyPrivate,
	})
	require.NoError(t, err)

	page1, err := svc.ListDiscover(ctx, memberID, "", 2, 0)
	require.NoError(t, err)
	require.Len(t, page1, 2)
	require.Equal(t, "Alpha Golf", page1[0].Group.Name)
	require.Equal(t, "Alpha Golf", page1[1].Group.Name)
	require.Less(t, page1[0].Group.ID, page1[1].Group.ID)

	page2, err := svc.ListDiscover(ctx, memberID, "", 2, 2)
	require.NoError(t, err)
	require.Len(t, page2, 2)
	require.Equal(t, "Mid Crew", page2[0].Group.Name)
	require.Equal(t, "Zebra Golf", page2[1].Group.Name)

	search, err := svc.ListDiscover(ctx, memberID, "alpha", 20, 0)
	require.NoError(t, err)
	require.Len(t, search, 2)
	for _, g := range search {
		require.Equal(t, "Alpha Golf", g.Group.Name)
		require.Equal(t, entity.GroupPrivacyPublic, g.Group.Privacy)
	}
}

func TestListMinePaginationAndOrder(t *testing.T) {
	svc, _ := newSvc()
	ctx := context.Background()
	for _, name := range []string{"Zulu", "Alpha", "Mike"} {
		g, err := svc.Create(ctx, port.CreateGroupInput{
			ActorID: ownerID, Name: name, Privacy: entity.GroupPrivacyPublic,
		})
		require.NoError(t, err)
		_, err = svc.Join(ctx, memberID, g.Group.ID)
		require.NoError(t, err)
	}

	page1, err := svc.ListMine(ctx, memberID, 2, 0)
	require.NoError(t, err)
	require.Len(t, page1, 2)
	require.Equal(t, "Alpha", page1[0].Group.Name)
	require.Equal(t, "Mike", page1[1].Group.Name)

	page2, err := svc.ListMine(ctx, memberID, 2, 2)
	require.NoError(t, err)
	require.Len(t, page2, 1)
	require.Equal(t, "Zulu", page2[0].Group.Name)
}
