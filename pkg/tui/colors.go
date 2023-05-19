package tui

import (
	"os"

	"github.com/muesli/termenv"
)

var (
	Blue = termenv.NewOutput(os.Stdout).Color("27")
)
