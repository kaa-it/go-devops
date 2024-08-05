package buildconfig

import (
	"bytes"
	"fmt"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func PrintBuildInfo() {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Build version: %s\n", buildVersion)
	fmt.Fprintf(&buf, "Build date: %s\n", buildDate)
	fmt.Fprintf(&buf, "Build commit: %s\n", buildCommit)
	fmt.Print(buf.String())
}
