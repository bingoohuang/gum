package confirm

import (
	"fmt"
	"os"

	"github.com/charmbracelet/gum/internal/exit"

	tea "github.com/charmbracelet/bubbletea"
)

// Run provides a shell script interface for prompting a user to confirm an
// action with an affirmative or negative answer.
func (o *Options) Run() error {
	m, err := tea.NewProgram(model{
		affirmative:      o.Affirmative,
		negative:         o.Negative,
		confirmation:     o.Default,
		defaultSelection: o.Default,
		timeout:          o.Timeout,
		hasTimeout:       o.Timeout > 0,
		prompt:           o.Prompt,
		selectedStyle:    o.SelectedStyle.ToLipgloss(),
		unselectedStyle:  o.UnselectedStyle.ToLipgloss(),
		promptStyle:      o.PromptStyle.ToLipgloss(),
	}, tea.WithOutput(os.Stderr)).Run()

	if err != nil {
		return fmt.Errorf("unable to run confirm: %w", err)
	}

	md := m.(model)
	if md.aborted {
		o.SetResult("ABORTED")
		if !o.AsAPI {
			os.Exit(exit.StatusAborted)
		}
		return nil
	}

	if md.confirmation {
		o.SetResult("YES")
		if !o.AsAPI {
			os.Exit(0)
		}
		return nil
	}

	o.SetResult("NO")
	if !o.AsAPI {
		os.Exit(1)
	}
	return nil
}
