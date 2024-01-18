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
func (o *Options) Run() error {
	for {
		if err := o.run(); err != nil {
			if errors.Is(err, exit.ErrAborted) {
				return nil
			}

			fmt.Println(err.Error())
		}
	}
}
func (o *Options) run() error {
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
	case "file":
		return o.setFile(words[1:])
	case "rm":
		return o.rmFile(words[1:])
	case "write":
		return o.writeFile(words[1:])
	case "read":
		return o.readFile(words[1:])
	case "append":
		return o.appendFile(words[1:])
	}

	return fmt.Errorf("unknown command %s", cmd)
}

func (o *Options) readFile(args []string) error {
	var fileName string
	f := flag.NewFlagSet("flag", flag.ExitOnError)
	f.StringVar(&fileName, "f", "", "file name")
	if err := f.Parse(args); err != nil {
		return err
	}

	var err error
	fileName, err = o.getFileName(fileName, f)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

func (o *Options) appendFile(args []string) error {
	var fileName string
	var text string
	f := flag.NewFlagSet("flag", flag.ExitOnError)
	f.StringVar(&fileName, "f", "", "file name")
	f.StringVar(&text, "t", "{{MovieName}}", "text")
	if err := f.Parse(args); err != nil {
		return err
	}

	var err error
	fileName, err = o.getFileName(fileName, f)
	if err != nil {
		return err
	}

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
	if _, err := fi.WriteString(text + "\n"); err != nil {
		return err
	}

	fmt.Printf("append %d bytes to %s successfully\n", len(text)+1, fileName)

	return nil
}

func (o *Options) setFile(args []string) error {
	var fileName string
	f := flag.NewFlagSet("flag", flag.ExitOnError)
	f.StringVar(&fileName, "f", "", "file name")
	if err := f.Parse(args); err != nil {
		return err
	}

	if fileName == "" && len(f.Args()) > 0 {
		fileName = f.Args()[0]
	}

	var err error
	fileName, err = homedir.Expand(fileName)
	if err != nil {
		return err
	}

	o.fileName = fileName
	return nil
}
func (o *Options) rmFile(args []string) error {
	var fileName string
	var err error

	f := flag.NewFlagSet("flag", flag.ExitOnError)
	f.StringVar(&fileName, "f", "", "file name")
	if err := f.Parse(args); err != nil {
		return err
	}

	fileName, err = o.getFileName(fileName, f)
	if err != nil {
		return err
	}

	return os.Remove(fileName)
}

func (o *Options) writeFile(args []string) error {
	var fileName string
	var text string
	var err error

	f := flag.NewFlagSet("flag", flag.ExitOnError)
	f.StringVar(&fileName, "f", "", "file name")
	f.StringVar(&text, "t", "{{MovieName}}", "text")
	if err := f.Parse(args); err != nil {
		return err
	}

	fileName, err = o.getFileName(fileName, f)
	if err != nil {
		return err
	}

	if text == "" {
		text = gofakeit.MovieName()
	} else if text, err = gofakeit.Template(text, nil); err != nil {
		return err
	}

	if err := os.WriteFile(fileName, []byte(text+"\n"), 0644); err != nil {
		return err
	}

	fmt.Printf("wrote %d bytes to %s successfully\n", len(text), fileName)
	return nil
}

func (o *Options) getFileName(fileName string, f *flag.FlagSet) (string, error) {
	if fileName == "" && len(f.Args()) > 0 {
		fileName = f.Args()[0]
	}

	var err error
	if fileName == "" {
		fileName = o.fileName
	} else {
		fileName, err = homedir.Expand(fileName)
		if err != nil {
			return "", err
		}
	}

	return fileName, nil
}
