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

func main() {
	cli.VersionFlag = &cli.BoolFlag{Name: "version", Aliases: []string{"V"}, Usage: "print the version"}

	cmd := &cli.Command{
		Name:    utils.Crac,
		Usage:   "container registry as cache",
		Version: utils.CracVersion.String(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "log-level", Value: "info",
				Usage: "log level, could be \"debug\", \"info\", \"warn\", \"error\"",
			},
			&cli.BoolFlag{
				Name: "silent", Aliases: []string{"s"},
				Usage: "disable all logging or stdout",
			},
		},
		Before: func(ctx context.Context, cli *cli.Command) (context.Context, error) {
			if cli.Bool("silent") {
				slog.SetDefault(slog.New(slog.DiscardHandler))
			} else {
				logLevel, err := func() (slog.Level, error) {
					switch l := cli.String("log-level"); l {
					case "debug":
						return slog.LevelDebug, nil
					case "info":
						return slog.LevelInfo, nil
					case "warn":
						return slog.LevelWarn, nil
					case "error":
						return slog.LevelError, nil
					default:
						return slog.LevelInfo, fmt.Errorf("log level \"%s\" is invalid", l)
					}
				}()
				if err != nil {
					return ctx, err
				}
				slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
				slog.SetLogLoggerLevel(logLevel)
			}
			return ctx, nil
		},
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
					&cli.StringFlag{
						Name: "tag", Aliases: []string{"t"}, Category: "BASIC",
						Usage: "specific a tag to push",
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
					&cli.BoolFlag{
						Name: "force", Category: "BASIC", Value: true,
						Usage: "force push to remote registry",
					},

					&cli.StringFlag{
						Name: "profile", Category: "PROFILE",
						Usage: "a series of pre-set configurations",
					},
					&cli.StringFlag{
						Name: "profile-file", Category: "PROFILE",
						Usage: "read profile from file, if set will ignore \"profile\"",
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
					repo := cmd.StringArg("repo")
					if len(repo) == 0 {
						return fmt.Errorf("argument repository is required")
					}
					deps := utils.ScanFiles(cmd.StringSlice("dep"))
					files := utils.ScanFiles(cmd.StringSlice("file"))
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

					output := cmd.String("output")

					platform := cmd.String("platform")
					if cmd.Bool("unknown-platform") {
						platform = "unknown/unknown"
					}

					return api.Push(
						api.WithContext(context.Background()),
						api.WithRepository(repo),
						api.WithUsername(cmd.String("username")),
						api.WithPassword(cmd.String("password")),
						api.WithForceHttp(cmd.Bool("force-http")),
						api.WithInsecure(cmd.Bool("insecure")),
						api.WithKeys(cmd.StringSlice("key")),
						api.WithDepFiles(deps),
						api.WithTag(cmd.String("tag")),
						api.WithFiles(files),
						api.WithWorkdir(cmd.String("workdir")),
						api.WithPlatform(platform),
						api.WithProfile(profile, func() string {
							if profileStdin {
								return "content"
							}
							if len(profileFile) != 0 {
								return "file"
							}
							return ""
						}()),
						api.WithOutputStdout(output == "stdout"),
						api.WithOutputFile(output),
						api.WithForcePush(cmd.Bool("force")),
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
					repo := cmd.StringArg("repo")
					if len(repo) == 0 {
						return fmt.Errorf("argument repository is required")
					}
					deps := utils.ScanFiles(cmd.StringSlice("dep"))
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
						api.WithKeys(cmd.StringSlice("key")),
						api.WithDepFiles(deps),
						api.WithTag(cmd.String("tag")),
						api.WithWorkdir(cmd.String("workdir")),
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
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
