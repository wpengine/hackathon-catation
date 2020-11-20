// Copyright (C) 2020  WPEngine
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

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

	"github.com/wpengine/hackathon-catation/internal"
	"github.com/wpengine/hackathon-catation/pup"
	"github.com/wpengine/hackathon-catation/pup/eternum"
	"github.com/wpengine/hackathon-catation/pup/pinata"
	"github.com/wpengine/hackathon-catation/pup/pipin"
)

/*
This command is here just to exercise the pup package for debugging.
*/

func main() {
	internal.PrintGPLBanner("catation", "2020")

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
			return ls(ctx, pinata.New(*pinataKey, *pinataSecret))
		},
	}

	pinataAdd := &ffcli.Command{
		Name:       "add",
		ShortUsage: "pup pinata add <hash>",
		Exec: func(ctx context.Context, args []string) error {
			return add(ctx, pinata.New(*pinataKey, *pinataSecret), args)
		},
	}

	pinataRm := &ffcli.Command{
		Name:       "rm",
		ShortUsage: "pup pinata rm <hash>",
		Exec: func(ctx context.Context, args []string) error {
			return rm(ctx, pinata.New(*pinataKey, *pinataSecret), args)
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
			return ls(ctx, pipin.New(true, *pipinHost, *pipinToken))
		},
	}

	pipinAdd := &ffcli.Command{
		Name:       "add",
		ShortUsage: "pup pipin add <hash>",
		Exec: func(ctx context.Context, args []string) error {
			return add(ctx, pipin.New(true, *pipinHost, *pipinToken), args)
		},
	}

	pipinRm := &ffcli.Command{
		Name:       "rm",
		ShortUsage: "pup pipin rm <hash>",
		Exec: func(ctx context.Context, args []string) error {
			return rm(ctx, pipin.New(true, *pipinHost, *pipinToken), args)
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
			return ls(ctx, eternum.New(*eternumKey))
		},
	}

	eternumAdd := &ffcli.Command{
		Name:       "add",
		ShortUsage: "pup eternum add <hash>",
		Exec: func(ctx context.Context, args []string) error {
			return add(ctx, eternum.New(*eternumKey), args)
		},
	}

	eternumRm := &ffcli.Command{
		Name:       "rm",
		ShortUsage: "pup eternum rm <hash>",
		Exec: func(ctx context.Context, args []string) error {
			return rm(ctx, eternum.New(*eternumKey), args)
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

func ls(ctx context.Context, client pup.Pup) error {
	hashes, err := client.Fetch(ctx, []pup.Hash{})
	if err != nil {
		return err
	}
	fmt.Println("Pinned Hashes")
	fmt.Println("----------------------------------------------")
	for _, hash := range hashes {
		fmt.Println(hash.Hash)
	}
	return nil
}

func add(ctx context.Context, client pup.Pup, args []string) error {
	if len(args) != 1 {
		return errors.New("add requires one hash argument")
	}
	if err := client.Pin(ctx, pup.Hash(args[0])); err != nil {
		return err
	}
	fmt.Printf("Pinned hash: %q\n", args[0])
	return nil
}

func rm(ctx context.Context, client pup.Pup, args []string) error {
	if len(args) != 1 {
		return errors.New("rm requires one hash argument")
	}
	if err := client.Unpin(ctx, pup.Hash(args[0])); err != nil {
		return err
	}
	fmt.Printf("Unpinned hash: %q\n", args[0])
	return nil
}
