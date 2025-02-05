package input

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/bingoohuang/gum/cursor"
	"github.com/bingoohuang/gum/internal/exit"
	"github.com/bingoohuang/gum/internal/stdin"
)

// Run provides a shell script interface for the text input bubble.
// https://github.com/charmbracelet/bubbles/textinput
func (o Options) Run() error {
	i := textinput.New()
	if o.Value != "" {
		i.SetValue(o.Value)
	} else if in, _ := stdin.Read(); in != "" {
		i.SetValue(in)
	}

	i.Focus()
	i.Prompt = o.Prompt
	i.Placeholder = o.Placeholder
	i.Width = o.Width
	i.PromptStyle = o.PromptStyle.ToLipgloss()
	i.Cursor.Style = o.CursorStyle.ToLipgloss()
	i.Cursor.SetMode(cursor.Modes[o.CursorMode])
	i.CharLimit = o.CharLimit

	if o.Password {
		i.EchoMode = textinput.EchoPassword
		i.EchoCharacter = '•'
	}

	p := tea.NewProgram(model{
		textinput:   i,
		aborted:     false,
		header:      o.Header,
		headerStyle: o.HeaderStyle.ToLipgloss(),
		timeout:     o.Timeout,
		hasTimeout:  o.Timeout > 0,
		autoWidth:   o.Width < 1,
	}, tea.WithOutput(os.Stderr))
	tm, err := p.Run()
	if err != nil {
		return fmt.Errorf("failed to run input: %w", err)
	}
	m := tm.(model)

	if m.aborted {
		return exit.ErrAborted
	}

	fmt.Println(m.textinput.Value())
	return nil
}
