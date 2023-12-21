package ps

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/gum/ansi"
	"github.com/charmbracelet/gum/choose"
	"github.com/charmbracelet/gum/internal/exit"
	"github.com/charmbracelet/gum/internal/utils"
	"github.com/creasty/defaults"
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
	o.Options = strings.Split(strings.TrimSuffix(input, "\n"), "\n")
	for i := 1; i < len(o.Options); i++ {
		fields := strings.Fields(o.Options[i])
		vsz, rss := fields[4], fields[5]
		vsz1, rss1 := humanizeBytes(vsz), humanizeBytes(rss)
		o.Options[i] = strings.ReplaceAll(o.Options[i], vsz, vsz1)
		o.Options[i] = strings.ReplaceAll(o.Options[i], rss, rss1)
	}
	options := []tea.ProgramOption{tea.WithOutput(os.Stderr)}
	if o.Height == 0 {
		options = append(options, tea.WithAltScreen())
	}

	var matches []fuzzy.Match
	if o.Value != "" {
		i.SetValue(o.Value)
	}
	switch {
	case o.Value != "":
		matches = append([]fuzzy.Match{{Str: o.Options[0]}}, exactMatches(o.Value, o.Options[1:])...)
	default:
		matches = matchAll(o.Options)
	}

	if o.NoLimit {
		o.Limit = len(o.Options)
	}

	v := viewport.New(o.Width, o.Height)
	p := tea.NewProgram(model{
		choices:               o.Options,
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
	}, options...)

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
	chooseOptions := &choose.Options{Options: []string{
		"kill",
		"kill -9",
		"kill -INT",
		"kill -HUP",
		"kill -USR1",
		"kill -USR2",
		"ignore",
	}}

	if err := defaults.Set(chooseOptions); err != nil {
		return err
	}
	if err := chooseOptions.Run(); err != nil {
		return err
	}

	fields := strings.Fields(o.GetResult())
	pid := fields[1]
	shell := ""

	result := chooseOptions.GetResult()
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
	return nil
}

func (o *Options) doResult(result string) {
	o.SetResult(result)
	fmt.Println(result)
}

func (o Options) checkSelected(m model, isTTY bool) {
	for k := range m.selected {
		if isTTY {
			fmt.Println(k)
		} else {
			fmt.Println(ansi.Strip(k))
		}
	}
}
