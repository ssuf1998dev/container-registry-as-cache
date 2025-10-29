package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ssuf1998dev/container-registry-as-cache/api"
	"github.com/urfave/cli/v3"
)

func scan(patterns []string) ([]string, error) {
	list := []string{}
	for _, item := range patterns {
		matches, err := filepath.Glob(item)
		if err != nil {
			return nil, err
		}
		list = append(list, matches...)
	}
	return list, nil
}

func main() {
	cli.VersionFlag = &cli.BoolFlag{Name: "version", Aliases: []string{"V"}, Usage: "print the version"}

	cmd := &cli.Command{
		Name:    api.Crac,
		Usage:   "container registry as cache",
		Version: api.CracVersion.String(),
		Commands: []*cli.Command{
			{
				Name:      "push",
				Usage:     "make cache image then push to the remote registry",
				ArgsUsage: "[repository]",
				Suggest:   false,
				Arguments: []cli.Argument{
					&cli.StringArg{Name: "repo"},
				},
				Flags: []cli.Flag{
					&cli.StringSliceFlag{Name: "keys", Usage: "key(s) for computing cache image tag", Aliases: []string{"k"}},
					&cli.StringSliceFlag{Name: "deps", Usage: "dependent file(s) for computing cache image tag, glob supported", Aliases: []string{"d"}},
					&cli.StringSliceFlag{Name: "files", Usage: "cache file(s) to make image, glob supported", Aliases: []string{"f"}},
					&cli.StringFlag{Name: "username", Usage: "username", Aliases: []string{"u"}},
					&cli.StringFlag{Name: "password", Usage: "password", Aliases: []string{"p"}},
					&cli.BoolFlag{Name: "insecure", Usage: "skip ssl verify for the remote registry", Aliases: []string{"i"}},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					repo := cmd.StringArg("repo")
					if len(repo) == 0 {
						return fmt.Errorf("argument \"repo\" is required")
					}
					deps, err := scan(cmd.StringSlice("deps"))
					if err != nil {
						return err
					}
					files, err := scan(cmd.StringSlice("files"))
					if err != nil {
						return err
					}

					return api.Push(
						api.WithContext(context.Background()),
						api.WithRepository(repo),
						api.WithUsername(cmd.String("username")),
						api.WithLoginUsername(),
						api.WithPassword(cmd.String("password")),
						api.WithLoginPassword(),
						api.WithInsecure(cmd.Bool("insecure")),
						api.WithKeys(cmd.StringSlice("keys")),
						api.WithDepFiles(deps),
						api.WithFiles(files),
					)
				},
			},
			{
				Name:      "pull",
				Usage:     "pull cache image then uncompress it",
				ArgsUsage: "[repository]",
				Suggest:   false,
				Arguments: []cli.Argument{
					&cli.StringArg{Name: "repo"},
				},
				Flags: []cli.Flag{
					&cli.StringSliceFlag{Name: "keys", Usage: "key(s) for computing cache image tag", Aliases: []string{"k"}},
					&cli.StringSliceFlag{Name: "deps", Usage: "dependent file(s) for computing cache image tag, glob supported", Aliases: []string{"d"}},
					&cli.StringFlag{Name: "tag", Usage: "specific a tag to pull", Aliases: []string{"t"}},
					&cli.StringFlag{Name: "workdir", Usage: "working directory where to uncompress file(s) to", Aliases: []string{"w"}},
					&cli.StringFlag{Name: "username", Usage: "username", Aliases: []string{"u"}},
					&cli.StringFlag{Name: "password", Usage: "password", Aliases: []string{"p"}},
					&cli.BoolFlag{Name: "insecure", Usage: "skip ssl verify for the remote registry", Aliases: []string{"i"}},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					repo := cmd.StringArg("repo")
					if len(repo) == 0 {
						return fmt.Errorf("argument \"repo\" is required")
					}
					deps, err := scan(cmd.StringSlice("deps"))
					if err != nil {
						return err
					}

					return api.Pull(
						api.WithContext(context.Background()),
						api.WithRepository(repo),
						api.WithUsername(cmd.String("username")),
						api.WithLoginUsername(),
						api.WithPassword(cmd.String("password")),
						api.WithLoginPassword(),
						api.WithInsecure(cmd.Bool("insecure")),
						api.WithKeys(cmd.StringSlice("keys")),
						api.WithDepFiles(deps),
						api.WithTag(cmd.String("tag")),
						api.WithWorkdir(cmd.String("workdir")),
					)
				},
			},
			{
				Name:      "login",
				Usage:     "authenticate to a registry",
				ArgsUsage: "[repository or registry]",
				Suggest:   false,
				Arguments: []cli.Argument{
					&cli.StringArg{Name: "repo"},
				},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "username", Usage: "username", Aliases: []string{"u"}},
					&cli.StringFlag{Name: "password", Usage: "password", Aliases: []string{"p"}},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					repo := cmd.StringArg("repo")
					if len(repo) == 0 {
						return fmt.Errorf("argument \"repo\" is required")
					}

					return api.Login(
						api.WithRepository(repo),
						api.WithUsername(cmd.String("username")),
						api.WithPassword(cmd.String("password")),
					)
				},
			},
			{
				Name:      "logout",
				Usage:     "log out from a registry",
				ArgsUsage: "[repository or registry]",
				Suggest:   false,
				Arguments: []cli.Argument{
					&cli.StringArg{Name: "repo"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					repo := cmd.StringArg("repo")
					if len(repo) == 0 {
						return fmt.Errorf("argument \"repo\" is required")
					}

					return api.Logout(
						api.WithRepository(repo),
					)
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
