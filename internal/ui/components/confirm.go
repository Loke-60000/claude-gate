package components

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ml0-1337/claude-gate/internal/ui/styles"
	"github.com/ml0-1337/claude-gate/internal/ui/utils"
)

// ConfirmModel represents a confirmation prompt
type ConfirmModel struct {
	question string
	answer   bool
	answered bool
}

// NewConfirm creates a new confirmation prompt
func NewConfirm(question string) ConfirmModel {
	return ConfirmModel{
		question: question,
		answer:   false,
		answered: false,
	}
}

// Init initializes the confirmation prompt
func (m ConfirmModel) Init() tea.Cmd {
	return nil
}

// Update handles confirmation updates
func (m ConfirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			m.answer = true
			m.answered = true
			return m, tea.Quit
		case "n", "N":
			m.answer = false
			m.answered = true
			return m, tea.Quit
		case "ctrl+c", "esc":
			m.answer = false
			m.answered = true
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the confirmation prompt
func (m ConfirmModel) View() string {
	if m.answered {
		answer := "No"
		if m.answer {
			answer = "Yes"
		}
		return fmt.Sprintf("%s %s\n", m.question, styles.InfoStyle.Render(answer))
	}
	return fmt.Sprintf("%s %s ", m.question, styles.HelpStyle.Render("(y/N)"))
}

// Confirm shows a confirmation prompt and returns the answer
func Confirm(question string) bool {
	// Check if we have a TTY available
	if !utils.IsInteractive() {
		return confirmNonInteractive(question, false)
	}

	model := NewConfirm(question)
	p := tea.NewProgram(model)
	
	finalModel, err := p.Run()
	if err != nil {
		return false
	}
	
	return finalModel.(ConfirmModel).answer
}

// ConfirmWithDefault shows a confirmation prompt with a default value
func ConfirmWithDefault(question string, defaultYes bool) bool {
	// Check if we have a TTY available
	if !utils.IsInteractive() {
		return confirmNonInteractive(question, defaultYes)
	}

	suffix := "(y/N)"
	if defaultYes {
		suffix = "(Y/n)"
	}
	
	fullQuestion := fmt.Sprintf("%s %s", question, styles.HelpStyle.Render(suffix))
	
	model := &ConfirmDefaultModel{
		ConfirmModel: ConfirmModel{
			question: fullQuestion,
			answer:   defaultYes,
			answered: false,
		},
		defaultYes: defaultYes,
	}
	
	p := tea.NewProgram(model)
	
	finalModel, err := p.Run()
	if err != nil {
		return false
	}
	
	return finalModel.(*ConfirmDefaultModel).answer
}

// confirmNonInteractive handles confirmation without TTY
func confirmNonInteractive(question string, defaultYes bool) bool {
	suffix := "(y/N)"
	if defaultYes {
		suffix = "(Y/n)"
	}
	
	fmt.Printf("%s %s ", question, suffix)
	
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		// On error or empty input, return the default
		return defaultYes
	}
	
	switch response {
	case "y", "Y", "yes", "Yes", "YES":
		return true
	case "n", "N", "no", "No", "NO":
		return false
	default:
		// Empty response or anything else, return default
		return defaultYes
	}
}

// ConfirmDefaultModel extends ConfirmModel with default value support
type ConfirmDefaultModel struct {
	ConfirmModel
	defaultYes bool
}

// Update handles confirmation updates with default support
func (m *ConfirmDefaultModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			m.answer = true
			m.answered = true
			return m, tea.Quit
		case "n", "N":
			m.answer = false
			m.answered = true
			return m, tea.Quit
		case "enter":
			m.answer = m.defaultYes
			m.answered = true
			return m, tea.Quit
		case "ctrl+c", "esc":
			m.answer = false
			m.answered = true
			return m, tea.Quit
		}
	}
	return m, nil
}