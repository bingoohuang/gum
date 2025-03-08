package tickwait

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/gum/bingoo"
	"github.com/charmbracelet/gum/internal/timeout"
)

type model struct {
	spinner spinner.Model

	result string
	done   bool
	start  time.Time
}

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Points
	return model{
		spinner: s,
		done:    false,
		start:   time.Now(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// 按 Ctrl+C 或 Esc 退出
		switch msg.Type {
		case tea.KeyCtrlC:
			m.result = "Ctrl+C"
			return m, tea.Interrupt
		case tea.KeyEsc:
			m.result = "ESC"
			return m, tea.Interrupt
		}

		if len(msg.String()) > 0 {
			m.result = msg.String()
			m.done = true
			return m, tea.Quit
		}

		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m model) View() string {
	cost := time.Since(m.start).Round(time.Millisecond)
	if m.done {
		return fmt.Sprintf("%s %s 等待完成", m.spinner.View(), cost)
	}

	return fmt.Sprintf("%s %s 按空格继续, Ctrl+C 或 Esc 或 超时 退出", m.spinner.View(), cost)
}

func (o Options) RunBingoo() (string, error) {
	ctx, cancel := timeout.Context(o.Timeout)
	defer cancel()

	m := initialModel()
	p := tea.NewProgram(m, tea.WithContext(ctx))
	tm, err := p.Run()
	m = tm.(model)
	cost := time.Since(m.start).Round(time.Millisecond)
	if err != nil {
		if o.TimeoutFn != nil && bingoo.IsErrorTimeout(err) {
			o.TimeoutFn(cost, m.spinner.View())
		}

		return "", err
	}

	if m.done {
		if o.DoneFn != nil {
			o.DoneFn(cost, m.spinner.View(), m.result)
		}
		return m.result, nil
	}

	return "", fmt.Errorf("error: %s", m.result)
}
