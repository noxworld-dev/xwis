package xwis

const (
	preHeaderLength  = 4
	headerLength     = 8
	fullHeaderLength = preHeaderLength + headerLength
)

func encryptTo(out, data []byte) []byte {
	ind := 0
	cnt := 0
	loc := 0
	loc2 := 0
	acc := 0
	for i := 0; i <= len(data)-10; i++ {
		for j := 0; j <= 7; j++ {
			if loc == 8 {
				ind++
				loc = 0
			}
			if loc2 == 7 {
				acc ^= 1 << 7
				out[cnt] = byte(acc)
				cnt++
				acc = 0
				loc2 = 0
			}
			var v6 byte
			if ind < len(data) {
				v6 = data[ind]
			}
			v5 := 1 << loc
			v4 := int(v6) & v5
			v3 := v4 >> loc
			acc ^= (v3 << loc2) & 0xFF
			loc++
			loc2++
		}
	}
	return out
}

func decrypt(data []byte) {
	ind := 0
	cnt := 0
	loc := 0
	for i := 0; i <= len(data)-10; i++ {
		var acc byte
		for j := 0; j <= 7; j++ {
			if cnt == 7 {
				cnt = 0
				loc++
			}
			if loc == 8 {
				data[ind] = 0
				ind++
				loc = 0
			}
			v6 := byte(0)
			if ind < len(data) {
				v6 = data[ind]
			}
			v5 := 1 << loc
			v4 := int(v6) & v5
			v3 := v4 >> loc
			acc ^= byte((v3 << j) & 0xFF)
			loc++
			cnt++
		}
		data[i] = acc
	}
}

func decryptAndDecode(data []byte) (*GameInfo, error) {
	data = data[fullHeaderLength:]
	decrypt(data)

	var g GameInfo
	err := g.UnmarshalBinary(data)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

var header = []byte{':', 'G', '1', 'P', '3', 0x9a, 0x03, 0x01}

func encodeAndEncrypt(g *GameInfo) ([]byte, error) {
	gdata, err := g.MarshalBinary()
	if err != nil {
		return nil, err
	}
	data := make([]byte, headerLength+len(gdata))
	copy(data, header)
	encryptTo(data[headerLength:], gdata)
	return data, nil
}
