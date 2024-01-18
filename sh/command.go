package sh

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/bingoohuang/gum/cursor"
	"github.com/bingoohuang/gum/internal/exit"
	"github.com/bingoohuang/gum/internal/stdin"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-shellwords"
	"github.com/mitchellh/go-homedir"
)

// Run provides a shell script interface for the text input bubble.
// https://github.com/charmbracelet/bubbles/textinput
func (o Options) Run() error {
	for {
		if err := o.run(); err != nil {
			if errors.Is(err, exit.ErrAborted) {
				return nil
			}

			fmt.Println(err.Error())
		}
	}
}
func (o Options) run() error {
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

	value := m.textinput.Value()
	fmt.Println(">", value)

	if value == "" {
		return nil
	}

	words, err := shellwords.Parse(value)
	if err != nil {
		return err
	}

	cmd := words[0]

	switch cmd {
	case "write":
		return writeFile(words[1:])
	case "read":
		return readFile(words[1:])
	case "append":
		return appendFile(words[1:])
	}

	return fmt.Errorf("unknown command %s", cmd)
}

func readFile(args []string) error {
	var fileName string
	f := flag.NewFlagSet("flag", flag.ExitOnError)
	f.StringVar(&fileName, "f", "", "file name")
	if err := f.Parse(args); err != nil {
		return err
	}

	if fileName == "" && len(f.Args()) > 0 {
		fileName = f.Args()[0]
	}
	fileName, _ = homedir.Expand(fileName)

	data, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

func appendFile(args []string) error {
	var fileName string
	var text string
	f := flag.NewFlagSet("flag", flag.ExitOnError)
	f.StringVar(&fileName, "f", "", "file name")
	f.StringVar(&text, "t", "{{MovieName}}", "text")
	if err := f.Parse(args); err != nil {
		return err
	}

	if fileName == "" && len(f.Args()) > 0 {
		fileName = f.Args()[0]
	}
	fileName, _ = homedir.Expand(fileName)

	fi, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer fi.Close()

	if text == "" {
		text = gofakeit.MovieName()
	} else if text, err = gofakeit.Template(text, nil); err != nil {
		return err
	}
	if _, err := fi.WriteString(text); err != nil {
		return err
	}

	fmt.Printf("append %d bytes to %s successfully\n", len(text), fileName)

	return nil
}

func writeFile(args []string) error {
	var fileName string
	var text string
	var err error

	f := flag.NewFlagSet("flag", flag.ExitOnError)
	f.StringVar(&fileName, "f", "", "file name")
	f.StringVar(&text, "t", "{{MovieName}}", "text")
	if err := f.Parse(args); err != nil {
		return err
	}

	if fileName == "" && len(f.Args()) > 0 {
		fileName = f.Args()[0]
	}

	fileName, _ = homedir.Expand(fileName)

	if text == "" {
		text = gofakeit.MovieName()
	} else if text, err = gofakeit.Template(text, nil); err != nil {
		return err
	}

	if err := os.WriteFile(fileName, []byte(text), 0644); err != nil {
		return err
	}

	fmt.Printf("wrote %d bytes to %s successfully\n", len(text), fileName)
	return nil
}
