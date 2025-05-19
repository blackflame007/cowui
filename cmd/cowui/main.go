package main

import (
	"fmt"
	"strings"

	cowsay "github.com/blackflame007/Neo-cowsay/v2"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the application state
type Model struct {
	messages    []string
	userInput   string
	responses   []string
	currentResp int
}

func getCowsay(message string) (string, error) {
	return cowsay.Say(
		message,
		cowsay.Type("bender"),
	)
}

// Initialize the model
func initialModel() Model {
	return Model{
		messages: []string{},
		responses: []string{
			"Bite my shiny metal ASCII! Type something to chat with me!",
			"Hey meatbag, nice to meet you!",
			"I'm 40% chatbot! *bangs chest*",
			"Wanna kill all humans? Just kidding... maybe.",
			"I need a drink... or ten!",
		},
		currentResp: 0,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.userInput == "" {
				return m, nil
			}

			// Get current Bender response
			benderResponse := m.responses[m.currentResp]

			// Add current Bender response to history first (if it's not already there)
			if len(m.messages) == 0 {
				m.messages = append(m.messages, "Bender: "+benderResponse)
			}

			// Add user message to history
			m.messages = append(m.messages, "You: "+m.userInput)

			// Advance to next response
			m.currentResp = (m.currentResp + 1) % len(m.responses)

			// Add next Bender response to history
			nextResponse := m.responses[m.currentResp]
			m.messages = append(m.messages, "Bender: "+nextResponse)

			// Clear input
			m.userInput = ""

			return m, nil
		case tea.KeyBackspace:
			if len(m.userInput) > 0 {
				m.userInput = m.userInput[:len(m.userInput)-1]
			}
		case tea.KeySpace:
			m.userInput += " "
		case tea.KeyRunes:
			m.userInput += string(msg.Runes)
		}
	}
	return m, nil
}

func (m Model) View() string {
	// Styles
	chatStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1).
		Width(60)

	cowStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1).
		Width(60)

	inputStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1).
		Width(60)

	// Chat history - show conversations in order, oldest at top
	var chatHistory strings.Builder
	for i := 0; i < len(m.messages); i++ {
		chatHistory.WriteString(m.messages[i] + "\n")
	}

	// Get current message for ASCII art
	var currentMessage string
	if len(m.messages) == 0 {
		// Initial state - show first response
		currentMessage = m.responses[0]
	} else {
		// Show the next response that will be used
		currentMessage = m.responses[m.currentResp]
	}

	// Generate Bender ASCII art using the neocowsay library
	benderArt, err := getCowsay(currentMessage)
	if err != nil {
		benderArt = fmt.Sprintf("Error generating cowsay: %v", err)
	}

	// Input prompt
	prompt := "Your message: " + m.userInput + "_"

	// Combine everything with proper spacing
	return lipgloss.JoinVertical(
		lipgloss.Left,
		chatStyle.Render(chatHistory.String()),
		cowStyle.Render(benderArt),
		inputStyle.Render(prompt),
		"\nPress Ctrl+C to exit",
	)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
	}
}
