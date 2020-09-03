package xwis

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClassString(t *testing.T) {
	require.Equal(t, "<none>", Class(0).String())
	require.Equal(t, "WIZ", ClassWizard.String())
	require.Equal(t, "WIZ | CON | Class(0x20)", (ClassWizard | ClassConjurer | 32).String())
}

func TestGameInfoDecode(t *testing.T) {
	var g GameInfo
	err := g.UnmarshalBinary([]byte(decodedInfo))
	require.NoError(t, err)
	require.Equal(t, GameInfo{
		Access:     AccessOpen,
		Disallow:   0,
		Unk1:       unk1Value,
		LimitRes:   false,
		Resolution: Res640x480,
		Players:    0,
		MaxPlayers: 29,
		MinPing:    -1,
		MaxPing:    -1,
		Unk2:       unk2Value,
		Map:        "headache",
		Name:       "NoxCommunity EU",
		Unk3:       unk3Data,
		Flags:      defaultFlags,
		MapType:    MapTypeArena,
		FragLimit:  15,
		TimeLimit:  0,
		Unknown:    make([]byte, unkLength),
	}, g)
}
