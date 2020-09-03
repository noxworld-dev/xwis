package main

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use: "list",
	}
	Root.AddCommand(cmd)
	fChats := cmd.Flags().Bool("chat", false, "list chat rooms")
	fInterval := cmd.Flags().Duration("t", time.Second*3, "refresh interval")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		rctx := cmd.Context()

		ctx, cancel := context.WithTimeout(rctx, time.Minute/2)
		cli, err := newClient(ctx)
		cancel()
		if err != nil {
			return err
		}
		defer cli.Close()

		cmd.SilenceUsage = true

		fmt.Println("Connected!")
		ticker := time.NewTicker(*fInterval)
		defer ticker.Stop()
		for {
			ctx, cancel := context.WithTimeout(rctx, time.Second*10)
			list, err := cli.ListRooms(ctx)
			cancel()
			if err != nil {
				return err
			}
			sort.Slice(list, func(i, j int) bool {
				return list[i].Name < list[j].Name
			})
			cnt := len(list)
			if !*fChats {
				cnt = 0
				for _, r := range list {
					if r.Game != nil {
						cnt++
					}
				}
			}
			fmt.Printf("\n%v\nTotal rooms: %d\n", time.Now().Format("2006-01-02 15:04:05"), cnt)
			for _, r := range list {
				if g := r.Game; g != nil {
					fmt.Printf("\t%s\t%d/%d\n", g.Name, g.Players, g.MaxPlayers)
				} else if *fChats {
					fmt.Printf("\t%s\t%d\n", r.Name, r.Users)
				}
			}
			select {
			case <-rctx.Done():
				return nil
			case <-ticker.C:
			}
		}
	}
}
