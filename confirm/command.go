package confirm

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/gum/internal/stdin"
	"github.com/charmbracelet/gum/internal/timeout"
	"github.com/charmbracelet/gum/style"
)

// RunBingoo provides a shell script interface for prompting a user to confirm an
// action with an affirmative or negative answer.
func (o Options) RunBingoo() (bool, error) {
	line, err := stdin.Read(stdin.SingleLine(true))
	if err == nil {
		switch line {
		case "yes", "y":
			return true, nil
		default:
			return false, nil
		}
	}

	ctx, cancel := timeout.Context(o.Timeout)
	defer cancel()

	top, right, bottom, left := style.ParsePadding(o.Padding)
	m := model{
		affirmative:      o.Affirmative,
		negative:         o.Negative,
		showOutput:       o.ShowOutput,
		confirmation:     o.Default,
		defaultSelection: o.Default,
		keys:             defaultKeymap(o.Affirmative, o.Negative),
		help:             help.New(),
		showHelp:         o.ShowHelp,
		prompt:           o.Prompt,
		promptFn:         o.PromptFn,
		selectedStyle:    o.SelectedStyle.ToLipgloss(),
		unselectedStyle:  o.UnselectedStyle.ToLipgloss(),
		promptStyle:      o.PromptStyle.ToLipgloss(),
		padding:          []int{top, right, bottom, left},
	}
	tm, err := tea.NewProgram(
		m,
		tea.WithOutput(os.Stderr),
		tea.WithContext(ctx),
	).Run()
	if err != nil && ctx.Err() != context.DeadlineExceeded {
		return false, fmt.Errorf("unable to confirm: %w", err)
	}
	m = tm.(model)

	if o.ShowOutput {
		confirmationText := m.negative
		if m.confirmation {
			confirmationText = m.affirmative
		}
		fmt.Println(m.getPrompt(), confirmationText)
	}

	return m.confirmation, nil
}
