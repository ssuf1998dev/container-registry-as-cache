package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/goccy/go-yaml"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/ssuf1998dev/container-registry-as-cache/internal/utils"
)

func Push(opts ...Option) error {
	o := &options{
		context: context.Background(),
	}
	for _, option := range opts {
		option(o)
	}

	_, err := push(o)
	return err
}

func push(opts *options) (image []byte, err error) {
	base := empty.Image

	if len(opts.files) == 0 {
		return nil, fmt.Errorf("empty image is not allowed")
	}

	slog.Info("making cache layer...", "files", len(opts.files))
	cacheLayer, err := utils.NewTarLayer(opts.files, opts.workdir)
	defer func() {
		os.Remove(cacheLayer.File)
	}()
	if err != nil {
		return nil, err
	}
	img, _ := mutate.AppendLayers(base, cacheLayer)
	slog.Info("cache layer done")

	meta, _ := yaml.Marshal(utils.CracMeta{
		Version: utils.CracVersion.String(),
	})
	metaLayer, _ := crane.Layer(map[string][]byte{fmt.Sprintf("/%s/meta.yaml", utils.Crac): meta})
	img, _ = mutate.AppendLayers(img, metaLayer)
	slog.Info("meta layer generated", "version", utils.CracVersion.String())

	cf, _ := img.ConfigFile()
	cf = cf.DeepCopy()
	cf.Created = v1.Time{Time: time.Now()}
	cf.History = []v1.History{
		{Created: cf.Created, CreatedBy: utils.CreatedByCracCopy},
		{Created: cf.Created, CreatedBy: utils.CreatedByCracMeta},
	}
	img, _ = mutate.ConfigFile(img, cf)

	var keys []string
	keys = append(keys, opts.keys...)
	if len(opts.platform) > 0 {
		keys = append(keys, opts.platform)
	}
	tag, err := utils.ComputeTag(opts.depFiles, keys, opts.workdir)
	if err != nil {
		return nil, err
	}
	repo := opts.repo
	if len(repo) == 0 {
		repo = fmt.Sprintf("%s/%s", name.DefaultRegistry, utils.Crac)
	}
	nameOpts := []name.Option{}
	if opts.forceHttp {
		nameOpts = append(nameOpts, name.Insecure)
	}
	ref, err := name.ParseReference(fmt.Sprintf("%s:%s", repo, tag), nameOpts...)
	if err != nil {
		return nil, err
	}
	slog.Info("reference", "repo", repo, "tag", tag, "keys", strings.Join(keys, ", "), "depFiles", len(opts.depFiles))
	imgSize, _ := img.Size()

	if opts.outputStdout {
		err := tarball.Write(ref, img, os.Stdout)
		return nil, err
	}

	if opts.outputBytes {
		var buf bytes.Buffer
		if err := tarball.Write(ref, img, &buf); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}

	if len(opts.outputFile) != 0 {
		slog.Info("writing the image to file...",
			"file", opts.outputFile,
			"bsize", imgSize,
			"size", humanize.Bytes(uint64(imgSize)),
		)
		f, err := os.OpenFile(opts.outputFile, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return nil, err
		}
		err = tarball.Write(ref, img, f)
		slog.Info("image wrote")
		return nil, err
	}

	transport := remote.DefaultTransport.(*http.Transport)
	if opts.insecure {
		transport = transport.Clone()
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	slog.Info("writing the image to remote registry...", "bsize", imgSize, "size", humanize.Bytes(uint64(imgSize)))
	remoteOpts := []remote.Option{
		remote.WithContext(opts.context),
		remote.WithTransport(transport),
	}
	if len(opts.username) != 0 && len(opts.password) != 0 {
		remoteOpts = append(remoteOpts, remote.WithAuth(&authn.Basic{Username: opts.username, Password: opts.password}))
	}
	err = remote.Write(ref, img, remoteOpts...)
	if err != nil {
		return nil, err
	}
	slog.Info("image wrote")
	return nil, nil
}
