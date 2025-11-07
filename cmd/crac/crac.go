package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/ssuf1998dev/container-registry-as-cache/api"
	"github.com/ssuf1998dev/container-registry-as-cache/internal/utils"
	"github.com/urfave/cli/v3"
)

func main() {
	cli.VersionFlag = &cli.BoolFlag{Name: "version", Aliases: []string{"V"}, Usage: "print the version"}

	cmd := &cli.Command{
		Name:    utils.Crac,
		Usage:   "container registry as cache",
		Version: utils.CracVersion.String(),
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
					&cli.StringSliceFlag{
						Name: "key", Aliases: []string{"K"}, Category: "BASIC",
						Usage: "key(s) for computing cache image tag",
					},
					&cli.StringSliceFlag{
						Name: "dep", Aliases: []string{"d"}, Category: "BASIC",
						Usage: "dependent file(s) for computing cache image tag, glob supported",
					},
					&cli.StringSliceFlag{
						Name: "file", Aliases: []string{"f"}, Category: "BASIC",
						Usage: "cache file(s) to make image, glob supported",
					},
					&cli.StringFlag{
						Name: "workdir", Aliases: []string{"w"}, Category: "BASIC",
						Usage: "working directory where to uncompress file(s) to",
					},
					&cli.StringFlag{
						Name: "platform", Aliases: []string{"P"}, Value: fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH), Category: "BASIC",
						Usage: "platform of cache, it will be a part of keys changing tag",
					},
					&cli.BoolFlag{
						Name: "unknown-platform", Category: "BASIC",
						Usage: "override platform of cache to unknown/unknown",
					},
					&cli.StringFlag{
						Name: "output", Aliases: []string{"o"}, Category: "BASIC",
						Usage: "output where, could be stdout or file",
					},

					&cli.StringFlag{
						Name: "profile", Category: "PROFILE",
						Usage: "a series of pre-set configurations",
					},
					&cli.StringFlag{
						Name: "profile-file", Category: "PROFILE",
						Usage: "read profile from file, if set will ignore \"profile\"",
					},

					&cli.StringFlag{
						Name: "username", Aliases: []string{"u"}, Category: "AUTH",
						Usage: "username for authenticating to a registry",
					},
					&cli.StringFlag{
						Name: "password", Aliases: []string{"p"}, Category: "AUTH",
						Usage: "password for authenticating to a registry",
					},
					&cli.BoolFlag{
						Name: "force-http", Category: "AUTH",
						Usage: "use http for the remote registry",
					},
					&cli.BoolFlag{
						Name: "insecure", Category: "AUTH",
						Usage: "skip ssl verify for the remote registry",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					repo := cmd.StringArg("repo")
					if len(repo) == 0 {
						return fmt.Errorf("argument repository is required")
					}
					deps, err := utils.ScanFiles(cmd.StringSlice("dep"))
					if err != nil {
						return err
					}
					files, err := utils.ScanFiles(cmd.StringSlice("file"))
					if err != nil {
						return err
					}
					profile := cmd.String("profile")
					profileFile := cmd.String("profile-file")
					if len(profileFile) != 0 {
						profile = profileFile
					}

					output := cmd.String("output")

					platform := cmd.String("platform")
					if cmd.Bool("unknown-platform") {
						platform = "unknown/unknown"
					}

					return api.Push(
						api.WithContext(context.Background()),
						api.WithRepository(repo),
						api.WithUsername(cmd.String("username")),
						api.WithLoginUsername(),
						api.WithPassword(cmd.String("password")),
						api.WithLoginPassword(),
						api.WithForceHttp(cmd.Bool("force-http")),
						api.WithInsecure(cmd.Bool("insecure")),
						api.WithKeys(cmd.StringSlice("key")),
						api.WithDepFiles(deps),
						api.WithFiles(files),
						api.WithWorkdir(cmd.String("workdir")),
						api.WithPlatform(platform),
						api.WithProfile(profile, len(profileFile) != 0),
						api.WithOutputStdout(output == "stdout"),
						api.WithOutputFile(output),
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
					&cli.StringSliceFlag{
						Name: "keys", Aliases: []string{"k"}, Category: "BASIC",
						Usage: "key(s) for computing cache image tag",
					},
					&cli.StringSliceFlag{
						Name: "deps", Aliases: []string{"d"}, Category: "BASIC",
						Usage: "dependent file(s) for computing cache image tag, glob supported",
					},
					&cli.StringFlag{
						Name: "tag", Aliases: []string{"t"}, Category: "BASIC",
						Usage: "specific a tag to pull",
					},
					&cli.StringFlag{
						Name: "workdir", Aliases: []string{"w"}, Category: "BASIC",
						Usage: "working directory where to uncompress file(s) to",
					},
					&cli.StringFlag{
						Name: "platform", Aliases: []string{"P"}, Value: fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH), Category: "BASIC",
						Usage: "platform of cache, it will be a part of keys changing tag",
					},
					&cli.BoolFlag{
						Name: "unknown-platform", Category: "BASIC",
						Usage: "override platform of cache to unknown/unknown",
					},
					&cli.BoolFlag{
						Name: "stdout", Category: "BASIC",
						Usage: "output to stdout",
					},

					&cli.StringFlag{
						Name: "profile", Category: "PROFILE",
						Usage: "a series of pre-set configurations",
					},
					&cli.StringFlag{
						Name: "profile-file", Category: "PROFILE",
						Usage: "read profile from file,  if set will ignore \"profile\"",
					},

					&cli.StringFlag{
						Name: "username", Aliases: []string{"u"}, Category: "AUTH",
						Usage: "username for authenticating to a registry",
					},
					&cli.StringFlag{
						Name: "password", Aliases: []string{"p"}, Category: "AUTH",
						Usage: "password for authenticating to a registry",
					},
					&cli.BoolFlag{
						Name: "force-http", Category: "AUTH",
						Usage: "use http for the remote registry",
					},
					&cli.BoolFlag{
						Name: "insecure", Category: "AUTH",
						Usage: "skip ssl verify for the remote registry",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					repo := cmd.StringArg("repo")
					if len(repo) == 0 {
						return fmt.Errorf("argument repository is required")
					}
					deps, err := utils.ScanFiles(cmd.StringSlice("deps"))
					if err != nil {
						return err
					}
					profile := cmd.String("profile")
					profileFile := cmd.String("profile-file")
					if len(profileFile) != 0 {
						profile = profileFile
					}

					platform := cmd.String("platform")
					if cmd.Bool("unknown-platform") {
						platform = "unknown/unknown"
					}

					return api.Pull(
						api.WithContext(context.Background()),
						api.WithRepository(repo),
						api.WithUsername(cmd.String("username")),
						api.WithLoginUsername(),
						api.WithPassword(cmd.String("password")),
						api.WithLoginPassword(),
						api.WithForceHttp(cmd.Bool("force-http")),
						api.WithInsecure(cmd.Bool("insecure")),
						api.WithKeys(cmd.StringSlice("keys")),
						api.WithDepFiles(deps),
						api.WithTag(cmd.String("tag")),
						api.WithWorkdir(cmd.String("workdir")),
						api.WithPlatform(platform),
						api.WithProfile(profile, len(profileFile) != 0),
						api.WithOutputStdout(cmd.Bool("stdout")),
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
