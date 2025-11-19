package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	cracprofile "github.com/ssuf1998dev/container-registry-as-cache/internal/profile"
	"github.com/ssuf1998dev/container-registry-as-cache/internal/utils"
	"github.com/urfave/cli/v3"
)

func stringSliceFlagRender(original []string, workdir string) []string {
	tpl := template.New("").Funcs(cracprofile.TplFuncs(workdir)).Funcs(sprig.FuncMap())

	results := make([]string, len(original))
	copy(results, original)

	for i, item := range original {
		parsed, err := tpl.Parse(item)
		if err != nil {
			continue
		}

		var buf bytes.Buffer
		err = parsed.Execute(&buf, nil)
		if err != nil {
			continue
		}

		results[i] = buf.String()
	}

	return results
}

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
			&push,
			&pull,
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
