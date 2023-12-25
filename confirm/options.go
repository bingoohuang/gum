package confirm

import (
	"github.com/charmbracelet/gum/internal/utils"
	"github.com/creasty/defaults"
	"time"

	"github.com/charmbracelet/gum/style"
)

// Options is the customization options for the confirm command.
type Options struct {
	Default     bool         `help:"Default confirmation action" default:"true"`
	Affirmative string       `help:"The title of the affirmative action" default:"Yes"`
	Negative    string       `help:"The title of the negative action" default:"No"`
	Prompt      string       `arg:"" help:"Prompt to display." default:"Are you sure?"`
	PromptStyle style.Styles `embed:"" prefix:"prompt." help:"The style of the prompt" set:"defaultMargin=1 0 0 0" envprefix:"GUM_CONFIRM_PROMPT_"`
	//nolint:staticcheck
	SelectedStyle style.Styles `embed:"" prefix:"selected." help:"The style of the selected action" set:"defaultBackground=212" set:"defaultForeground=230" set:"defaultPadding=0 3" set:"defaultMargin=1 1" envprefix:"GUM_CONFIRM_SELECTED_"`
	//nolint:staticcheck
	UnselectedStyle style.Styles  `embed:"" prefix:"unselected." help:"The style of the unselected action" set:"defaultBackground=235" set:"defaultForeground=254" set:"defaultPadding=0 3" set:"defaultMargin=1 1" envprefix:"GUM_CONFIRM_UNSELECTED_"`
	Timeout         time.Duration `help:"Timeout until confirm returns selected value or default if provided" default:"0" env:"GUM_CONFIRM_TIMEOUT"`

	utils.Result

	AsAPI bool // 是否以 API 的方式调用，非(GUM 本身的方式），不会直接 print，或者 exit
}

// Confirm API方式确认
func (o *Options) Confirm() (string, error) {
	o.AsAPI = true
	if err := defaults.Set(o); err != nil {
		return "", err
	}

	if err := o.Run(); err != nil {
		return "", err
	}

	return o.GetResult(), nil
}
