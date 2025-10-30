package utils

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

var Crac = "crac"
var CreatedByCracMeta = "CRACMETA"
var CreatedByCracCopy = "CRACCOPY"
var CracVersion = semver.MustParse("1.0.0")
var CracVersionConstraint, _ = semver.NewConstraint(fmt.Sprintf(">= %d < %d", CracVersion.Major(), CracVersion.Major()+1))
