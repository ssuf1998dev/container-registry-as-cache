package main

import (
	"bufio"
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"runtime"

	"github.com/ssuf1998dev/container-registry-as-cache/api"
	"github.com/ssuf1998dev/container-registry-as-cache/internal/utils"
	"github.com/urfave/cli/v3"
)

var pull = cli.Command{
	Name:      "pull",
	Usage:     "pull cache image then uncompress it",
	ArgsUsage: "[repository]",
	Suggest:   false,
	Arguments: []cli.Argument{
		&cli.StringArg{Name: "repo"},
	},
	Flags: []cli.Flag{
		&cli.StringSliceFlag{
			Name: "key", Aliases: []string{"k"}, Category: "BASIC",
			Usage: "key(s) for computing cache image tag",
		},
		&cli.StringSliceFlag{
			Name: "dep", Aliases: []string{"d"}, Category: "BASIC",
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
		&cli.Uint32Flag{
			Name: "perm", Category: "BASIC", Value: 0755, DefaultText: "0755",
			Usage: "chmod all pulled file",
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
		&cli.BoolFlag{
			Name: "profile-stdin", Category: "PROFILE",
			Usage: "read profile from stdin, if set will ignore \"profile\" and \"profile-file\"",
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
		slog.Debug("pulling...")

		workdir := cmd.String("workdir")

		repo := cmd.StringArg("repo")
		if len(repo) == 0 {
			return fmt.Errorf("argument repository is required")
		}

		keys := stringSliceFlagRender(cmd.StringSlice("key"), workdir)
		deps := utils.ScanFiles(stringSliceFlagRender(cmd.StringSlice("dep"), workdir))
		profile := cmd.String("profile")
		profileFile := cmd.String("profile-file")
		profileStdin := cmd.Bool("profile-stdin")
		if profileStdin {
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				profile += fmt.Sprintf("%s\n", scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				return err
			}
		} else if len(profileFile) != 0 {
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
			api.WithPassword(cmd.String("password")),
			api.WithForceHttp(cmd.Bool("force-http")),
			api.WithInsecure(cmd.Bool("insecure")),
			api.WithKeys(keys),
			api.WithDepFiles(deps),
			api.WithTag(cmd.String("tag")),
			api.WithWorkdir(workdir),
			api.WithPlatform(platform),
			api.WithFilePerm(fs.FileMode(cmd.Uint32("perm"))),
			api.WithProfile(profile, func() string {
				if profileStdin {
					return "content"
				}
				if len(profileFile) != 0 {
					return "file"
				}
				return ""
			}()),
			api.WithOutputStdout(cmd.Bool("stdout")),
		)
	},
}
