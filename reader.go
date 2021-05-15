package xwis

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"gopkg.in/irc.v3"
)

var DebugLog *log.Logger

type reader struct {
	sc *bufio.Scanner
}

func newReader(r io.Reader) *reader {
	return &reader{sc: bufio.NewScanner(r)}
}

func (r *reader) ReadLine() (string, error) {
	if r.sc.Scan() {
		return r.sc.Text(), nil
	}
	if err := r.sc.Err(); err != nil {
		return "", err
	}
	return "", io.EOF
}

func (r *reader) ReadMessage() (*irc.Message, error) {
	line, err := r.ReadLine()
	if err != nil {
		return nil, err
	}
	return irc.ParseMessage(line)
}

func (r *reader) WaitFor(ctx context.Context, cmds ...string) (*irc.Message, error) {
	done := ctx.Done()
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
			return nil, fmt.Errorf(pkg+": wait(%s): %w", strings.Join(cmds, "|"), err)
		}
		for _, c := range cmds {
			if c == m.Command {
				return m, nil
			}
		}
		if DebugLog != nil {
			DebugLog.Println(m)
		}
	}
}
