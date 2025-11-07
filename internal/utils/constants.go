package utils

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
)

//go:embed version.txt
var version string

var Crac = "crac"
var CreatedByCracMeta = "CRACMETA"
var CreatedByCracCopy = "CRACCOPY"
var CracVersion = semver.MustParse(strings.TrimSpace(version))
var CracVersionConstraint, _ = semver.NewConstraint(fmt.Sprintf(">= %d < %d", CracVersion.Major(), CracVersion.Major()+1))
