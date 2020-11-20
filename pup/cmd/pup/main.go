package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/wpengine/hackathon-catation/pup"
	"github.com/wpengine/hackathon-catation/pup/eternum"
	"github.com/wpengine/hackathon-catation/pup/pinata"
	"github.com/wpengine/hackathon-catation/pup/pipin"
)

/*
This command is here just to exercise the pup package for debugging.
*/

func main() {

	var (
		pipinFlags = flag.NewFlagSet("pup pipin", flag.ExitOnError)
		pipinHost  = pipinFlags.String("host", "pipin.velvetcache.org", "PiPin hostname")
		pipinToken = pipinFlags.String("token", "", "PiPin authentication token")

		pinataFlags  = flag.NewFlagSet("pup pinata", flag.ExitOnError)
		pinataKey    = pinataFlags.String("api-key", "", "Pinata service API key")
		pinataSecret = pinataFlags.String("secret-api-key", "", "Pinata service secret API key")

		eternumFlags = flag.NewFlagSet("pup eternum", flag.ExitOnError)
		eternumKey   = eternumFlags.String("api-key", "", "Eternum API key")
	)

	//////////////////////////////////////////////////////////
	// Pinata

	pinataList := &ffcli.Command{
		Name:       "ls",
		ShortUsage: "pup pinata ls",
		Exec: func(ctx context.Context, args []string) error {
			client := pinata.New(*pinataKey, *pinataSecret)
			client.Fetch(ctx, []pup.Hash{})
			hashes, err := client.Fetch(ctx, []pup.Hash{})
			if err != nil {
				return err
			}
			fmt.Println("Pinned Hashes")
			fmt.Println("-------------")
			for _, hash := range hashes {
				fmt.Println(hash.Hash)
			}
			return nil
		},
	}

	pinataAdd := &ffcli.Command{
		Name:       "add",
		ShortUsage: "pup pinata add <hash>",
		Exec: func(ctx context.Context, args []string) error {
			if len(args) != 1 {
				return errors.New("add requires one hash argument")
			}
			client := pinata.New(*pinataKey, *pinataSecret)
			err := client.Pin(ctx, pup.Hash(args[0]))
			if err != nil {
				return err
			}
			fmt.Printf("Pinned hash: %q\n", args[0])
			return nil
		},
	}

	pinataRm := &ffcli.Command{
		Name:       "rm",
		ShortUsage: "pup pinata rm <hash>",
		Exec: func(ctx context.Context, args []string) error {
			if len(args) != 1 {
				return errors.New("rm requires one hash argument")
			}
			client := pinata.New(*pinataKey, *pinataSecret)
			err := client.Unpin(ctx, pup.Hash(args[0]))
			if err != nil {
				return err
			}
			fmt.Printf("Unpinned hash: %q\n", args[0])
			return nil
		},
	}

	pinataRoot := &ffcli.Command{
		Name:        "pinata",
		ShortUsage:  "pup pinata [flags] <command>",
		FlagSet:     pinataFlags,
		Options:     []ff.Option{ff.WithEnvVarPrefix("PINATA")},
		Subcommands: []*ffcli.Command{pinataList, pinataAdd, pinataRm},
	}

	/////////////////////////////////////////////////////////
	// PiPin

	pipinList := &ffcli.Command{
		Name:       "ls",
		ShortUsage: "pup pipin ls",
		Exec: func(ctx context.Context, args []string) error {
			client := pipin.New(true, *pipinHost, *pipinToken)
			hashes, err := client.Fetch(ctx, []pup.Hash{})
			if err != nil {
				return err
			}
			fmt.Println("Pinned Hashes")
			fmt.Println("-------------")
			for _, hash := range hashes {
				fmt.Println(hash.Hash)
			}
			return nil
		},
	}

	pipinAdd := &ffcli.Command{
		Name:       "add",
		ShortUsage: "pup pipin add <hash>",
		Exec: func(ctx context.Context, args []string) error {
			if len(args) != 1 {
				return errors.New("add requires one hash argument")
			}
			client := pipin.New(true, *pipinHost, *pipinToken)
			err := client.Pin(ctx, pup.Hash(args[0]))
			if err != nil {
				return err
			}
			fmt.Printf("Pinned hash: %q\n", args[0])
			return nil
		},
	}

	pipinRm := &ffcli.Command{
		Name:       "rm",
		ShortUsage: "pup pipin rm <hash>",
		Exec: func(ctx context.Context, args []string) error {
			if len(args) != 1 {
				return errors.New("rm requires one hash argument")
			}
			client := pipin.New(true, *pipinHost, *pipinToken)
			err := client.Unpin(ctx, pup.Hash(args[0]))
			if err != nil {
				return err
			}
			fmt.Printf("Unpinned hash: %q\n", args[0])
			return nil
		},
	}

	pipinRoot := &ffcli.Command{
		Name:        "pipin",
		ShortUsage:  "pup pipin [flags] <command>",
		FlagSet:     pipinFlags,
		Options:     []ff.Option{ff.WithEnvVarPrefix("PIPIN")},
		Subcommands: []*ffcli.Command{pipinList, pipinAdd, pipinRm},
	}

	/////////////////////////////////////////////////////////
	// Eternum

	eternumList := &ffcli.Command{
		Name:       "ls",
		ShortUsage: "pup eternum ls",
		Exec: func(ctx context.Context, args []string) error {
			client := eternum.New(*eternumKey)
			hashes, err := client.Fetch(ctx, []pup.Hash{})
			if err != nil {
				return err
			}
			fmt.Println("Pinned Hashes")
			fmt.Println("-------------")
			for _, hash := range hashes {
				fmt.Println(hash.Hash)
			}
			return nil
		},
	}

	eternumAdd := &ffcli.Command{
		Name:       "add",
		ShortUsage: "pup eternum add <hash>",
		Exec: func(ctx context.Context, args []string) error {
			if len(args) != 1 {
				return errors.New("add requires one hash argument")
			}
			client := eternum.New(*eternumKey)
			err := client.Pin(ctx, pup.Hash(args[0]))
			if err != nil {
				return err
			}
			fmt.Printf("Pinned hash: %q\n", args[0])
			return nil
		},
	}

	eternumRm := &ffcli.Command{
		Name:       "rm",
		ShortUsage: "pup eternum rm <hash>",
		Exec: func(ctx context.Context, args []string) error {
			if len(args) != 1 {
				return errors.New("rm requires one hash argument")
			}
			client := eternum.New(*eternumKey)
			err := client.Unpin(ctx, pup.Hash(args[0]))
			if err != nil {
				return err
			}
			fmt.Printf("Unpinned hash: %q\n", args[0])
			return nil
		},
	}

	eternumRoot := &ffcli.Command{
		Name:        "eternum",
		ShortUsage:  "pup eternum [flags] <command>",
		FlagSet:     eternumFlags,
		Options:     []ff.Option{ff.WithEnvVarPrefix("ETERNUM")},
		Subcommands: []*ffcli.Command{eternumList, eternumAdd, eternumRm},
	}

	/////////////////////////////////////////////////////////

	root := &ffcli.Command{
		ShortUsage:  "pup [flags] <command>",
		Subcommands: []*ffcli.Command{pipinRoot, pinataRoot, eternumRoot},
	}

	if err := root.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
