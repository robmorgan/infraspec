package build

import (
	"runtime/debug"
)

// Version is dynamically set by the toolchain or overridden by the Makefile.
var Version = "DEV"

// Commit is dynamically set by the toolchain or overridden in the Makefile.
var Commit = ""

// Date is dynamically set by the toolchain or overridden in the Makefile.
var Date = "" // YYYY-MM-DD

func init() {
	if Version == "DEV" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "(devel)" {
			Version = info.Main.Version
		}
	}
}
