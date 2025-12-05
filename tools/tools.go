//go:build tools
// +build tools

package tools

// This file tracks CLI-only tool dependencies via blank imports.
import (
	_ "github.com/bitfield/gotestdox/cmd/gotestdox"
)

//go:generate go install github.com/bitfield/gotestdox/cmd/gotestdox@latest
