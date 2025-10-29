package api

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

var CRAC_VERSION = semver.MustParse("1")
var CRAC_VERSION_CONSTRAINT, _ = semver.NewConstraint(fmt.Sprintf(">= %d < %d", CRAC_VERSION.Major(), CRAC_VERSION.Major()+1))

type BaseOptions struct {
	Repo     string
	Username string
	Password string
	Insecure bool
}

type Meta struct {
	Version string `json:"created,omitempty"`
}
