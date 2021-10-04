package xwis

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/irc.v3"
)

var (
	dialer     net.Dialer
	rander     = rand.New(rand.NewSource(time.Now().UnixNano()))
	lobbyNames = []string{
		"Brin",
		"Ix",
		"Dun Mir",
	}
)

var (
	ErrClientClosed = errors.New("client closed")
)

const (
	pkg            = "xwis"
	maxLogin       = 9
	DefaultAddress = "xwis.net:4000"
	defaultTimeout = time.Minute / 2
)

type LobbyServer struct {
	Addr string
	Name string
}

func randomLogin() string {
	return fmt.Sprintf("probe%04x", rander.Intn(0x10000))
}

func ListLobbyServers(ctx context.Context) ([]LobbyServer, error) {
	conn, err := dialer.DialContext(ctx, "tcp", DefaultAddress)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	name := randomLogin()
	return listLobbyServers(ctx, conn, name)
}

func listLobbyServers(ctx context.Context, conn net.Conn, name string) ([]LobbyServer, error) {
	w := newWriter(conn)
	if err := w.WriteLine("verchk 32512 65551"); err != nil {
		return nil, err
	}
	if err := w.WriteLine("verchk 9472 65540"); err != nil {
		return nil, err
	}
	if err := w.WriteLine("lobcount 9472"); err != nil {
		return nil, err
	}
	if err := w.WriteLinef("whereto %s %s 9472 65540 2227973051451322323085", name, name); err != nil {
		return nil, err
	}
	if err := w.WriteLine("QUIT"); err != nil {
		return nil, err
	}
	if err := w.Flush(); err != nil {
		return nil, err
	}
	done := ctx.Done()

	var out []LobbyServer
	r := newReader(conn)
	for {
		select {
		case <-done:
			return nil, ctx.Err()
		default:
		}
		m, err := r.ReadMessage()
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		if err != nil {
			return nil, fmt.Errorf(pkg+": %w", err)
		}
		switch m.Command {
		case "610":
			// ignore
		case "607":
			return out, nil
		case "605":
			if len(m.Params) < 2 {
				return nil, fmt.Errorf(pkg+": unexpected line: %q", m.String())
			}
			args := strings.SplitN(m.Params[1], " ", 4)
			if len(args) != 4 {
				return nil, fmt.Errorf(pkg+": unexpected line: %q", m.String())
			}
			addr := strings.Join([]string{
				args[0], // addr
				args[1], // port
			}, ":")
			name := strings.Trim(args[2], "'")
			if sname := strings.SplitN(name, ":", 2); len(sname) > 1 {
				name = sname[1]
			}
			out = append(out, LobbyServer{
				Addr: addr,
				Name: name,
			})
		}
	}
}

func NewClient(ctx context.Context, login, pass string) (*Client, error) {
	return NewClientWithAddress(ctx, DefaultAddress, login, pass)
}

func NewClientWithAddress(ctx context.Context, addr, login, pass string) (*Client, error) {
	if login == "" {
		login = randomLogin()
	}
	if pass == "" {
		pass = login
	}
	if len(login) > maxLogin {
		login = login[:maxLogin]
	}
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	c := &Client{
		c:     conn,
		w:     newWriter(conn),
		r:     newReader(conn),
		login: login,
		stop:  make(chan struct{}),
	}
	if err := c.handshake(ctx, host, pass); err != nil {
		_ = conn.Close()
		return nil, err
	}
	go c.readLoop()
	return c, nil
}

type readStream struct {
	c    *Client
	stop chan struct{}
	msg  chan *irc.Message
	errc chan error
}

func (s *readStream) WaitFor(ctx context.Context, cmds ...string) (*irc.Message, error) {
	done := ctx.Done()
	for {
		var m *irc.Message
		select {
		case <-s.c.stop:
			return nil, ErrClientClosed
		case <-done:
			return nil, ctx.Err()
		case err := <-s.errc:
			return nil, fmt.Errorf(pkg+": wait(%s): %w", strings.Join(cmds, "|"), err)
		case m = <-s.msg:
		}
		for _, c := range cmds {
			if c == m.Command {
				return m, nil
			}
		}
	}
}

func (s *readStream) Close() error {
	s.c.mu.Lock()
	defer s.c.mu.Unlock()
	if s.c.read == s {
		s.c.read = nil
	}
	close(s.stop)
	return nil
}

type Client struct {
	login string
	mu    sync.Mutex
	c     net.Conn
	w     *writer
	r     *reader // owned by readLoop
	stop  chan struct{}
	read  *readStream
}

func (c *Client) curStream() *readStream {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.read
}

func (c *Client) readLoop() {
	for {
		select {
		case <-c.stop:
			return
		default:
		}
		m, err := c.r.ReadMessage()
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		if err != nil {
			err = fmt.Errorf(pkg+": %w", err)
			if DebugLog != nil {
				DebugLog.Println(err)
			}
			if s := c.curStream(); s != nil {
				select {
				case <-c.stop:
				case <-s.stop:
				case s.errc <- err:
				}
			}
			return
		}
		if DebugLog != nil {
			DebugLog.Println(m)
		}
		if s := c.curStream(); s != nil {
			select {
			case <-c.stop:
			case <-s.stop:
			case s.msg <- m:
			}
		}
	}
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	select {
	case <-c.stop:
		return nil
	default:
	}
	defer close(c.stop)
	_ = c.c.SetWriteDeadline(time.Now().Add(time.Second))
	_ = c.w.WriteLine("QUIT")
	_ = c.w.Flush()
	return c.c.Close()
}

func getDeadline(ctx context.Context) time.Time {
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(defaultTimeout)
	}
	return deadline
}

func (c *Client) handshake(ctx context.Context, host, pass string) error {
	const (
		versCheck = false
		setCP     = false
		setOpt    = false
	)
	c.mu.Lock()
	defer c.mu.Unlock()
	deadline := getDeadline(ctx)
	if err := c.c.SetWriteDeadline(deadline); err != nil {
		return err
	}
	defer c.c.SetWriteDeadline(time.Time{})
	if err := c.w.WriteLine("CVERS 11015 9472"); err != nil {
		return err
	}
	if err := c.w.WriteLine("PASS supersecret"); err != nil {
		return err
	}
	if err := c.w.WriteLinef("NICK %s", c.login); err != nil {
		return err
	}
	if err := c.w.WriteLinef("apgar %s 0", pass); err != nil {
		return err
	}
	if err := c.w.WriteLinef("USER UserName HostName %s :RealName", host); err != nil {
		return err
	}
	if versCheck {
		if err := c.w.WriteLine("verchk 32512 720911"); err != nil {
			return err
		}
	}
	if setOpt {
		if err := c.w.WriteLine("SETOPT 17,33"); err != nil {
			return err
		}
	}
	if setCP {
		if err := c.w.WriteLine("SETCODEPAGE 1252"); err != nil {
			return err
		}
	}
	if err := c.w.Flush(); err != nil {
		return err
	}

	if err := c.c.SetReadDeadline(deadline); err != nil {
		return err
	}
	defer c.c.SetReadDeadline(time.Time{})

	// login itself
	if _, err := c.r.WaitFor(ctx, "376"); err != nil {
		return err
	}

	if versCheck {
		if _, err := c.r.WaitFor(ctx, "379"); err != nil {
			return err
		}
	}
	if setCP {
		if _, err := c.r.WaitFor(ctx, "329"); err != nil {
			return err
		}
	}
	return nil
}

type Room struct {
	ID    string
	Name  string
	Users int
	Game  *GameInfo
}

func (c *Client) newStreamUnsafe() *readStream {
	c.read = &readStream{
		c:    c,
		stop: make(chan struct{}),
		msg:  make(chan *irc.Message, 5),
		errc: make(chan error, 1),
	}
	return c.read
}

func (c *Client) lockWhenAvailable(ctx context.Context) error {
	for {
		c.mu.Lock()
		s := c.read
		if s == nil {
			return nil // still holding mutex
		}
		c.mu.Unlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.stop:
			return ErrClientClosed
		case <-s.stop:
		}
	}
}

func (c *Client) writeListRoomsReq(ctx context.Context) (*readStream, error) {
	if err := c.lockWhenAvailable(ctx); err != nil {
		return nil, err
	}
	defer c.mu.Unlock()
	deadline := getDeadline(ctx)
	if err := c.c.SetDeadline(deadline); err != nil {
		return nil, err
	}
	defer c.c.SetDeadline(time.Time{})
	// TODO: 37 is probably a game ID (Nox)
	if err := c.w.WriteLine("LIST -1 37"); err != nil {
		return nil, err
	}
	if err := c.w.Flush(); err != nil {
		return nil, err
	}
	return c.newStreamUnsafe(), nil
}

// ListRooms lists all available rooms on XWIS.
func (c *Client) ListRooms(ctx context.Context) ([]Room, error) {
	read, err := c.writeListRoomsReq(ctx)
	if err != nil {
		return nil, err
	}
	defer read.Close()
	var out []Room
	for {
		var m *irc.Message
		select {
		case <-c.stop:
			return nil, ErrClientClosed
		case err := <-read.errc:
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			return nil, err
		case m = <-read.msg:
		}
		switch m.Command {
		case "326": // game
			if len(m.Params) != 9 {
				return nil, fmt.Errorf(pkg+": unexpected line: %q", m.String())
			}
			id := m.Params[1]
			name := strings.TrimPrefix(id, "#")
			payload := m.Params[8]
			info, err := decryptAndDecode([]byte(payload))
			if err != nil {
				log.Printf("cannot parse game info: %v", err)
			} else {
				if v, err := strconv.ParseUint(m.Params[7], 10, 32); err != nil {
					log.Printf("cannot parse game addr: %v", err)
				} else {
					ip := net.IPv4(byte(v>>24), byte(v>>16), byte(v>>8), byte(v>>0))
					info.Addr = ip.String()
				}
			}
			r := Room{
				ID:   id,
				Name: name,
				Game: info,
			}
			if info != nil {
				r.Name = info.Name
				r.Users = info.Players
			}
			out = append(out, r)
		case "327": // chat
			if len(m.Params) != 5 {
				return nil, fmt.Errorf(pkg+": unexpected line: %q", m.String())
			}
			id := m.Params[1]
			name := strings.TrimPrefix(id, "#")
			if strings.HasPrefix(name, "Lob_37_") {
				ind, err := strconv.ParseUint(name[7:], 10, 8)
				if err == nil && ind >= 0 && int(ind) < len(lobbyNames) {
					name = lobbyNames[ind]
				}
			}
			num, err := strconv.ParseUint(m.Params[2], 10, 16)
			if err != nil {
				return nil, fmt.Errorf(pkg+": %w", err)
			}
			out = append(out, Room{
				ID:    id,
				Name:  name,
				Users: int(num),
			})
		case "323":
			return out, nil
		}
	}
}

func (c *Client) writeNewChannelReq(ctx context.Context, info *GameInfo) (string, *readStream, error) {
	if err := c.lockWhenAvailable(ctx); err != nil {
		return "", nil, err
	}
	defer c.mu.Unlock()
	channel := fmt.Sprintf("#%s's_game", c.login)
	if err := c.w.WriteLinef("JOINGAME %s 1 %d 37 3 1 1 13893824", channel, info.MaxPlayers); err != nil {
		return "", nil, err
	}
	if err := c.w.Flush(); err != nil {
		return "", nil, err
	}
	return channel, c.newStreamUnsafe(), nil
}

func (c *Client) writeStartGameReq(ctx context.Context, channel string, info *GameInfo) error {
	payload, err := encodeAndEncrypt(info)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.w.WriteLinef("TOPIC %s %s", channel, string(payload)); err != nil {
		return err
	}
	if false {
		if err := c.w.WriteLinef("STARTG %s %s", channel, c.login); err != nil {
			return err
		}
		if err := c.w.WriteLinef("TOPIC %s %s", channel, string(payload)); err != nil {
			return err
		}
	}
	if err := c.w.Flush(); err != nil {
		return err
	}
	return nil
}

func (c *Client) writeUpdateGameReq(ctx context.Context, channel string, info *GameInfo) error {
	payload, err := encodeAndEncrypt(info)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.w.WriteLinef("TOPIC %s %s", channel, string(payload)); err != nil {
		return err
	}
	if err := c.w.Flush(); err != nil {
		return err
	}
	return nil
}

func (c *Client) writeHostGameReq(ctx context.Context, info *GameInfo) (string, error) {
	channel, read, err := c.writeNewChannelReq(ctx, info)
	if err != nil {
		return "", err
	}
	if _, err := read.WaitFor(ctx, "366"); err != nil {
		return "", err
	}
	_ = read.Close()
	if err := c.writeStartGameReq(ctx, channel, info); err != nil {
		return "", err
	}
	return channel, nil
}

func (c *Client) writeStopGameReq(channel string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.w.WriteLinef("PART %s", channel)
	return c.w.Flush()
}

// HostGame registers a game and keeps it online until the context is cancelled.
// This call blocks for the whole duration of the game.
func (c *Client) HostGame(ctx context.Context, info GameInfo) error {
	g, err := c.RegisterGame(ctx, info)
	if err != nil {
		return err
	}
	defer g.Close()
	select {
	case <-c.stop:
		return ErrClientClosed
	case <-ctx.Done():
	}
	return nil
}

type Game struct {
	c       *Client
	info    GameInfo
	channel string
	closed  bool
}

// Update info for this game.
func (g *Game) Update(ctx context.Context, info GameInfo) error {
	info.setDefaults()
	return g.c.writeUpdateGameReq(ctx, g.channel, &info)
}

// Close the game and remove it from XWIS.
func (g *Game) Close() error {
	if g.closed {
		return nil
	}
	g.closed = true
	return g.c.writeStopGameReq(g.channel)
}

// RegisterGame register the game online and allows to control it asynchronously.
func (c *Client) RegisterGame(ctx context.Context, info GameInfo) (*Game, error) {
	info.setDefaults()
	channel, err := c.writeHostGameReq(ctx, &info)
	if err != nil {
		return nil, err
	}
	return &Game{c: c, info: info, channel: channel}, nil
}
