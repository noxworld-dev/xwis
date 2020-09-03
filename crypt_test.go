package xwis

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	encodedInfoHdr = "128::G1P3\x9a\x03\x01\x80\xfe\x83\x80\xd0\xe3\xff\xff\xff\xff\xfbĄ\xadٰ\xe4\u008d\xc3\u058c\x80\xa7\xef\xf0\x8d\xfa֭ۺ\xee\xd2\xd1ˇ\xa4Ѫ\xff\xff\xff\xff\xff\xff\xfb\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\x87¼\x80\x80\x80"
	decodedInfo    = "\x00\xff\x00\x00\x1d\xff\xff\xff\xff\x9eHheadache\x00NoxCommunity EU\xff\xff\xff\xff\xff\xef\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\a!\x0f\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"
)

func TestDecrypt(t *testing.T) {
	data := []byte(encodedInfoHdr[fullHeaderLength:])
	decrypt(data)
	require.Equal(t, decodedInfo, string(data))
}

func TestEncrypt(t *testing.T) {
	data := []byte(decodedInfo)
	out := make([]byte, len(data))
	encryptTo(out, data)
	require.Equal(t, encodedInfoHdr[fullHeaderLength:], string(out))
}

func TestDecodeEncode(t *testing.T) {
	g, err := decryptAndDecode([]byte(encodedInfoHdr))
	require.NoError(t, err)
	data, err := encodeAndEncrypt(g)
	require.NoError(t, err)
	require.Equal(t, encodedInfoHdr[preHeaderLength:], string(data))
}
