package bingoo

import (
	"errors"

	"github.com/alecthomas/kong"
	tea "github.com/charmbracelet/bubbletea"
)

// KongParse constructs a new parser and parses the default command-line.
func KongParse(cli any, options ...kong.Option) *kong.Context {
	parser, err := kong.New(cli, options...)
	if err != nil {
		panic(err)
	}
	ctx, err := parser.Parse(nil)
	parser.FatalIfErrorf(err)
	return ctx
}

var KongVars = kong.Vars{
	"defaultHeight":           "0",
	"defaultWidth":            "0",
	"defaultAlign":            "left",
	"defaultBorder":           "none",
	"defaultBorderForeground": "",
	"defaultBorderBackground": "",
	"defaultBackground":       "",
	"defaultForeground":       "",
	"defaultMargin":           "0 0",
	"defaultPadding":          "0 0",
	"defaultUnderline":        "false",
	"defaultBold":             "false",
	"defaultFaint":            "false",
	"defaultItalic":           "false",
	"defaultStrikethrough":    "false",
}

func IsErrorTimeout(err error) bool {
	return errors.Is(err, tea.ErrProgramKilled)
}

func IsErrorAborted(err error) bool {
	return errors.Is(err, tea.ErrInterrupted)
}
