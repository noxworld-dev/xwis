package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/noxworld-dev/xwis"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use: "register",
	}
	Root.AddCommand(cmd)
	fConf := cmd.Flags().StringP("config", "c", "xwis-game.json", "game config")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		rctx := cmd.Context()

		data, err := ioutil.ReadFile(*fConf)
		if os.IsNotExist(err) {
			cmd.SilenceUsage = true
			g := xwis.GameInfo{
				Access:     xwis.AccessOpen,
				Resolution: xwis.Res640x480,
				Players:    0,
				MaxPlayers: 31,
				Map:        "mymap",
				Name:       "My Server",
				MapType:    xwis.MapTypeArena,
				FragLimit:  15,
			}
			data, err = json.MarshalIndent(g, "", "\t")
			if err != nil {
				return err
			}
			err = ioutil.WriteFile(*fConf, data, 0644)
			if err != nil {
				return err
			}
			return errors.New("config not found - generated a new one; please edit and re-run the command")
		}
		if err != nil {
			return err
		}
		var g xwis.GameInfo
		if err := json.Unmarshal(data, &g); err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(rctx, time.Minute/2)
		cli, err := newClient(ctx)
		cancel()
		if err != nil {
			return err
		}
		defer cli.Close()

		cmd.SilenceUsage = true

		fmt.Printf("Hosting game: %q on %q (%s)\n", g.Name, g.Map, g.MapType)
		return cli.HostGame(rctx, g)
	}
}
