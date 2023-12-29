package ps

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/bingoohuang/gum/ansi"
	"github.com/bingoohuang/gum/choose"
	"github.com/bingoohuang/gum/internal/exit"
	"github.com/bingoohuang/gum/internal/utils"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dustin/go-humanize"
	"github.com/mattn/go-isatty"
	"github.com/sahilm/fuzzy"
)

// Run provides a shell script interface for filtering through options, powered
// by the textinput bubble.
func (o *Options) Run() error {
	i := textinput.New()
	i.Focus()

	i.Prompt = o.Prompt
	i.PromptStyle = o.PromptStyle.ToLipgloss()
	i.Placeholder = o.Placeholder
	i.Width = o.Width

	input, _, err := utils.Shellout("ps aux")
	if err != nil {
		return err
	}
	options := strings.Split(strings.TrimSuffix(input, "\n"), "\n")
	for i := 1; i < len(options); i++ {
		fields := strings.Fields(options[i])
		vsz, rss := fields[4], fields[5]
		vsz1, rss1 := humanizeBytes(vsz), humanizeBytes(rss)
		options[i] = strings.ReplaceAll(options[i], vsz, vsz1)
		options[i] = strings.ReplaceAll(options[i], rss, rss1)
	}
	teaOptions := []tea.ProgramOption{tea.WithOutput(os.Stderr)}
	if o.Height == 0 {
		teaOptions = append(teaOptions, tea.WithAltScreen())
	}

	var matches []fuzzy.Match
	if o.Value == "" && len(o.Options) > 0 {
		o.Value = o.Options[0]
	}
	if o.Value != "" {
		i.SetValue(o.Value)
	}
	switch {
	case o.Value != "":
		matches = append([]fuzzy.Match{{Str: options[0]}}, exactMatches(o.Value, options[1:])...)
	default:
		matches = matchAll(options)
	}

	if o.NoLimit {
		o.Limit = len(options) - 1
	}

	v := viewport.New(o.Width, o.Height)
	p := tea.NewProgram(model{
		choices:               options,
		indicator:             o.Indicator,
		matches:               matches,
		header:                o.Header,
		textinput:             i,
		viewport:              &v,
		indicatorStyle:        o.IndicatorStyle.ToLipgloss(),
		selectedPrefixStyle:   o.SelectedPrefixStyle.ToLipgloss(),
		selectedPrefix:        o.SelectedPrefix,
		unselectedPrefixStyle: o.UnselectedPrefixStyle.ToLipgloss(),
		unselectedPrefix:      o.UnselectedPrefix,
		matchStyle:            o.MatchStyle.ToLipgloss(),
		headerStyle:           o.HeaderStyle.ToLipgloss(),
		textStyle:             o.TextStyle.ToLipgloss(),
		cursorTextStyle:       o.CursorTextStyle.ToLipgloss(),
		height:                o.Height,
		selected:              make(map[string]struct{}),
		limit:                 o.Limit,
		reverse:               o.Reverse,
		timeout:               o.Timeout,
		hasTimeout:            o.Timeout > 0,
		sort:                  o.Sort,
	}, teaOptions...)

	tm, err := p.Run()
	if err != nil {
		return fmt.Errorf("unable to run filter: %w", err)
	}
	m := tm.(model)
	if m.aborted {
		return exit.ErrAborted
	}

	isTTY := isatty.IsTerminal(os.Stdout.Fd())

	// allSelections contains values only if limit is greater
	// than 1 or if flag --no-limit is passed, hence there is
	// no need to further checks
	if len(m.selected) > 0 {
		o.checkSelected(m, isTTY)
	} else if len(m.matches) > m.cursor && m.cursor >= 0 {
		if isTTY {
			o.doResult(m.matches[m.cursor].Str)
		} else {
			o.doResult(ansi.Strip(m.matches[m.cursor].Str))
		}
	}

	if !o.Strict && len(m.textinput.Value()) != 0 && len(m.matches) == 0 {
		o.doResult(m.textinput.Value())
	}

	if err := o.dealPsOutput(); err != nil {
		if errors.Is(err, ErrIgnore) {
			return nil
		}
		return err
	}

	o.Value = m.TextInputValue()
	return o.Run()
}

func humanizeBytes(sizeStr string) string {
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		return sizeStr
	}

	bytes := humanize.Bytes(uint64(size))
	bytes = strings.ReplaceAll(bytes, " ", "")
	diff := len(sizeStr) - len(bytes)
	if diff < 0 {
		return sizeStr
	}

	return strings.Repeat(" ", diff) + bytes
}

var ErrIgnore = errors.New("ignored")

func (o *Options) dealPsOutput() error {
	chooseOptions := &choose.Options{}
	kong.Parse(chooseOptions, kong.Vars{
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
	})
	chooseOptions.Options = []string{
		"kill",
		"kill -9",
		"kill -INT",
		"kill -HUP",
		"kill -USR1",
		"kill -USR2",
		"ignore",
	}

	if err := chooseOptions.Run(); err != nil {
		return err
	}

	result := chooseOptions.GetResult0()

	for _, psResult := range o.GetResult() {
		fields := strings.Fields(psResult)
		pid := fields[1]
		shell := ""

		switch {
		case strings.Contains(result, "kill -9"):
			shell = "kill -9 " + pid
		case strings.Contains(result, "kill"):
			shell = strings.TrimSpace(result) + " " + pid
		case result == "ignore":
			return ErrIgnore
		}

		if shell == "" {
			return nil
		}

		fmt.Printf("shell: %q\n", shell)
		stdOut, stdErr, err := utils.Shellout(shell)
		if err != nil {
			return err
		}
		if stdOut != "" {
			fmt.Print(stdOut)
		}
		if stdErr != "" {
			os.Stderr.WriteString(stdErr)
		}
	}

	return nil
}

func (o *Options) doResult(result string) {
	o.SetResult(result)
	fmt.Println(result)
}

func (o *Options) checkSelected(m model, isTTY bool) {
	var result []string
	for k := range m.selected {
		result = append(result, k)
		if isTTY {
			fmt.Println(k)
		} else {
			fmt.Println(ansi.Strip(k))
		}
	}

	o.SetResult(result...)
}
