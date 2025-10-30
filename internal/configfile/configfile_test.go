package configfile

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWrite(t *testing.T) {
	opts := &NewOptions{File: "../../testdata/config.json"}
	cf := NewConfigFile(opts)
	err := cf.Read()
	require.NoError(t, err)

	_, err = os.Stat(opts.File)
	require.NoError(t, err)

	if cf.Config.Auths == nil {
		cf.Config.Auths = map[string]ConfigAuth{}
	}
	cf.Config.Auths["foo"] = ConfigAuth{
		Username: "testuser",
		Password: "testpassword",
	}

	err = cf.Write()
	require.NoError(t, err)

	err = cf.Reset()
	require.NoError(t, err)
}
