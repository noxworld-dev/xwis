package xwis

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHost(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	const servUser = "testserv"
	info := GameInfo{
		Access:     AccessOpen,
		Resolution: Res640x480,
		Players:    3,
		MaxPlayers: 32,
		Map:        "headache",
		Name:       "Test Server",
		MapType:    MapTypeChat,
		FragLimit:  5,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	srv, err := NewClient(ctx, servUser, servUser)
	require.NoError(t, err)
	defer srv.Close()

	err = srv.HostGame(ctx, info)
	require.NoError(t, err)
}

func TestList(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	srv, err := NewClient(ctx, "", "")
	require.NoError(t, err)
	defer srv.Close()

	list, err := srv.ListRooms(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, list)
	for _, r := range list {
		if r.Game != nil {
			t.Logf("%q - %s", r.Name, r.Game.Addr)
		} else {
			t.Logf("%q", r.Name)
		}
	}
}

func TestListLobbyServers(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	list, err := ListLobbyServers(ctx)
	require.NoError(t, err)
	require.Equal(t, []LobbyServer{
		{Addr: "xwis.net:4000", Name: "XWIS"},
	}, list)
}
