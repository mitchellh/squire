package cli

import (
	"github.com/fatih/color"
)

var (
	colorSuccess     = color.New(color.FgGreen)
	colorError       = color.New(color.FgRed, color.Bold)
	colorErrorDetail = color.New(color.FgRed)
)
