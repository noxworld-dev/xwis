package xwis

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	timeLimitRes = time.Second // TODO
	unk1Value    = 0xff
	unk2Value    = 0x489e
	unkLength    = 9
	defaultFlags = 8199
)

var unk3Data = [28]byte{
	0xff, 0xff, 0xff, 0xff, 0xff, 0xef, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
}

var (
	endiness                            = binary.LittleEndian
	_        encoding.BinaryMarshaler   = &GameInfo{}
	_        encoding.BinaryUnmarshaler = &GameInfo{}

	_ json.Marshaler   = Access(0)
	_ json.Unmarshaler = (*Access)(nil)

	_ json.Marshaler   = Resolution(0)
	_ json.Unmarshaler = (*Resolution)(nil)

	_ json.Marshaler   = MapType(0)
	_ json.Unmarshaler = (*MapType)(nil)
)

const (
	AccessOpen    = Access(0)
	AccessClosed  = Access(1)
	AccessPrivate = Access(2)
)

type Access int

func (a Access) Unknown() bool {
	switch a {
	case AccessOpen, AccessClosed, AccessPrivate:
		return false
	}
	return true
}

func (a Access) String() string {
	switch a {
	case AccessOpen:
		return "open"
	case AccessClosed:
		return "closed"
	case AccessPrivate:
		return "private"
	}
	return fmt.Sprintf("Access(%d)", int(a))
}

func (a Access) MarshalJSON() ([]byte, error) {
	if a.Unknown() {
		return json.Marshal(int(a))
	}
	return json.Marshal(a.String())
}

func (a *Access) UnmarshalJSON(data []byte) error {
	var v int
	err := json.Unmarshal(data, &v)
	if err == nil {
		*a = Access(v)
		return nil
	}
	var s string
	err = json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	switch s {
	case "open", "public":
		*a = AccessOpen
	case "closed":
		*a = AccessClosed
	case "private":
		*a = AccessPrivate
	default:
		return fmt.Errorf("unsupported access value: %q", s)
	}
	return nil
}

const (
	ClassWarrior  = Class(1 << 0)
	ClassWizard   = Class(1 << 1)
	ClassConjurer = Class(1 << 2)
)

type Class int

func (c Class) String() string {
	var arr []string
	if c&ClassWarrior != 0 {
		arr = append(arr, "WAR")
		c &^= ClassWarrior
	}
	if c&ClassWizard != 0 {
		arr = append(arr, "WIZ")
		c &^= ClassWizard
	}
	if c&ClassConjurer != 0 {
		arr = append(arr, "CON")
		c &^= ClassConjurer
	}
	if c != 0 {
		arr = append(arr, fmt.Sprintf("Class(0x%x)", int(c)))
	}
	if len(arr) == 0 {
		return "<none>"
	}
	return strings.Join(arr, " | ")
}

const (
	Res640x480   = Resolution(0)
	Res800x600   = Resolution(1)
	Res1024x768  = Resolution(2)
	Res1280x1024 = Resolution(3)
)

type Resolution int

func (r Resolution) Unknown() bool {
	switch r {
	case Res640x480, Res800x600, Res1024x768, Res1280x1024:
		return false
	}
	return true
}

func (r Resolution) String() string {
	switch r {
	case Res640x480:
		return "640x480"
	case Res800x600:
		return "800x600"
	case Res1024x768:
		return "1024x768"
	case Res1280x1024:
		return "1280x1024"
	}
	return fmt.Sprintf("Resolution(%d)", int(r))
}

func (r Resolution) MarshalJSON() ([]byte, error) {
	if r.Unknown() {
		return json.Marshal(int(r))
	}
	return json.Marshal(r.String())
}

func (r *Resolution) UnmarshalJSON(data []byte) error {
	var v int
	err := json.Unmarshal(data, &v)
	if err == nil {
		*r = Resolution(v)
		return nil
	}
	var s string
	err = json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	switch s {
	case "640x480":
		*r = Res640x480
	case "800x600":
		*r = Res800x600
	case "1024x768":
		*r = Res1024x768
	case "1280x1024":
		*r = Res1280x1024
	default:
		return fmt.Errorf("unsupported resolution value: %q", s)
	}
	return nil
}

const (
	MapTypeKOTR        = MapType(0x0010)
	MapTypeCTF         = MapType(0x0020)
	MapTypeFlagBall    = MapType(0x0040)
	MapTypeChat        = MapType(0x0080)
	MapTypeArena       = MapType(0x0100)
	MapTypeElimination = MapType(0x0400)
	MapTypeCoop        = MapType(0x0A00)
	MapTypeQuest       = MapType(0x1000)
)

type GameFlags int

type MapType int

func (m MapType) Unknown() bool {
	switch m {
	case MapTypeKOTR, MapTypeCTF, MapTypeFlagBall, MapTypeChat,
		MapTypeArena, MapTypeElimination, MapTypeCoop, MapTypeQuest:
		return false
	}
	return true
}

func (m MapType) String() string {
	switch m {
	case MapTypeKOTR:
		return "kotr"
	case MapTypeCTF:
		return "ctf"
	case MapTypeFlagBall:
		return "flagball"
	case MapTypeChat:
		return "chat"
	case MapTypeArena:
		return "arena"
	case MapTypeElimination:
		return "elimination"
	case MapTypeCoop:
		return "coop"
	case MapTypeQuest:
		return "quest"
	}
	return fmt.Sprintf("MapType(0x%x)", int(m))
}

func (m MapType) MarshalJSON() ([]byte, error) {
	if m.Unknown() {
		return json.Marshal(int(m))
	}
	return json.Marshal(m.String())
}

func (m *MapType) UnmarshalJSON(data []byte) error {
	var v int
	err := json.Unmarshal(data, &v)
	if err == nil {
		*m = MapType(v)
		return nil
	}
	var s string
	err = json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	switch s {
	case "kotr":
		*m = MapTypeKOTR
	case "ctf":
		*m = MapTypeCTF
	case "flagball":
		*m = MapTypeFlagBall
	case "chat":
		*m = MapTypeChat
	case "arena":
		*m = MapTypeArena
	case "elimination":
		*m = MapTypeElimination
	case "coop":
		*m = MapTypeCoop
	case "quest":
		*m = MapTypeQuest
	default:
		return fmt.Errorf("unsupported map type value: %q", s)
	}
	return nil
}

type GameInfo struct {
	Addr       string        `json:"addr"`
	Name       string        `json:"name"`
	Map        string        `json:"map"`
	MapType    MapType       `json:"map_type"`
	Access     Access        `json:"access"`
	Disallow   Class         `json:"disallow,omitempty"`
	Flags      GameFlags     `json:"flags,omitempty"`
	Resolution Resolution    `json:"resolution"`
	LimitRes   bool          `json:"limit_res,omitempty"`
	Players    int           `json:"players,omitempty"`
	MaxPlayers int           `json:"max_players,omitempty"`
	MinPing    int           `json:"min_ping,omitempty"`
	MaxPing    int           `json:"max_ping,omitempty"`
	FragLimit  int           `json:"frag_limit,omitempty"`
	TimeLimit  time.Duration `json:"time_limit,omitempty"`
	Unk1       byte          `json:"-"`
	Unk2       uint16        `json:"-"`
	Unk3       [28]byte      `json:"-"`
	Unknown    []byte        `json:"-"`
}

func (g *GameInfo) setDefaults() {
	if g.MaxPlayers == 32 {
		g.MaxPlayers--
	}
	if g.Unk1 == 0 {
		g.Unk1 = unk1Value
	}
	if g.Unk2 == 0 {
		g.Unk2 = unk2Value
	}
	if g.Unk3 == ([28]byte{}) {
		copy(g.Unk3[:], unk3Data[:])
	}
	if len(g.Unknown) == 0 {
		g.Unknown = make([]byte, unkLength)
	}
	if g.Flags == 0 {
		g.Flags = defaultFlags
	}
}

func (g *GameInfo) MarshalBinary() ([]byte, error) {
	data := make([]byte, 69+len(g.Unknown))
	p := data

	// byte 0: access code
	p[0] = (byte(g.Access<<4) & 0xF0) | (byte(g.Disallow) & 0x0F)
	p = p[1:]

	// byte 1: unknown
	p[0] = g.Unk1
	p = p[1:]

	// byte 2: screen resolution code
	if g.LimitRes {
		p[0] = 8 << 4
	}
	p[0] |= byte(g.Resolution) & 0xF
	p = p[1:]

	// byte 3: players
	p[0] = byte(g.Players)
	p = p[1:]

	// byte 4: max players
	p[0] = byte(g.MaxPlayers)
	p = p[1:]

	// byte 5-6: min ping
	v16 := uint16(g.MinPing)
	if g.MinPing <= 0 {
		v16 = math.MaxUint16
	}
	endiness.PutUint16(p, v16)
	p = p[2:]

	// byte 7-8: max ping
	v16 = uint16(g.MaxPing)
	if g.MaxPing <= 0 {
		v16 = math.MaxUint16
	}
	endiness.PutUint16(p, v16)
	p = p[2:]

	// byte 9-10: unknown
	endiness.PutUint16(p, g.Unk2)
	p = p[2:]

	// byte 11-19: map name
	copy(p[:9], g.Map)
	p = p[9:]

	// byte 20-34: game name
	copy(p[:15], g.Name)
	p = p[15:]

	// byte 35-62: unknown
	copy(p[:28], g.Unk3[:])
	p = p[28:]

	// byte 63-64: game flags
	v16 = uint16(g.Flags) | uint16(g.MapType)
	endiness.PutUint16(p, v16)
	p = p[2:]

	// byte 65-66: frag limit
	endiness.PutUint16(p, uint16(g.FragLimit))
	p = p[2:]

	// byte 67-68: time limit
	endiness.PutUint16(p, uint16(g.TimeLimit/timeLimitRes))
	p = p[2:]

	copy(p, g.Unknown)
	return data, nil
}

func (g *GameInfo) UnmarshalBinary(data []byte) error {
	*g = GameInfo{}

	// byte 0: access code
	acc := data[0]
	data = data[1:]

	g.Access = Access((acc & 0xF0) >> 4)
	g.Disallow = Class(acc & 0x0F)

	// byte 1: unknown
	g.Unk1 = data[0]
	data = data[1:]

	// byte 2: screen resolution code
	res := data[0]
	data = data[1:]

	g.LimitRes = (res & (8 << 4)) != 0
	g.Resolution = Resolution(res & 0xF)

	// byte 3: players
	g.Players = int(data[0])
	data = data[1:]

	// byte 4: max players
	g.MaxPlayers = int(data[0])
	data = data[1:]

	// byte 5-6: min ping
	g.MinPing = int(endiness.Uint16(data))
	data = data[2:]

	if g.MinPing == math.MaxUint16 {
		g.MinPing = -1
	}

	// byte 7-8: max ping
	g.MaxPing = int(endiness.Uint16(data))
	data = data[2:]

	if g.MaxPing == math.MaxUint16 {
		g.MaxPing = -1
	}

	// byte 9-10: unknown
	g.Unk2 = endiness.Uint16(data)
	data = data[2:]

	// byte 11-19: map name
	mname := data[:9]
	data = data[9:]

	if i := bytes.IndexByte(mname, 0); i >= 0 {
		mname = mname[:i]
	}
	g.Map = string(mname)

	// byte 20-34: game name
	gname := data[:15]
	data = data[15:]

	if i := bytes.IndexByte(gname, 0); i >= 0 {
		gname = gname[:i]
	}
	g.Name = string(gname)

	// byte 35-62: unknown
	copy(g.Unk3[:], data[:28])
	data = data[28:]

	// byte 63-64: game flags
	g.Flags = GameFlags(endiness.Uint16(data))
	data = data[2:]

	g.MapType = MapType(g.Flags & 0x1FF0)
	g.Flags &^= GameFlags(0x1FF0)

	// byte 65-66: frag limit
	g.FragLimit = int(endiness.Uint16(data))
	data = data[2:]

	// byte 67-68: time limit
	g.TimeLimit = time.Duration(endiness.Uint16(data)) * timeLimitRes
	data = data[2:]

	g.Unknown = append([]byte{}, data...)

	return nil
}
