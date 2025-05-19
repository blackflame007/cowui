package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	cowsay "github.com/blackflame007/Neo-cowsay/v2"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
)

const (
	apiBaseURL = "http://localhost:3000/api"
)

// Model represents the application state
type Model struct {
	messages    []string
	userInput   string
	agentID     string
	loading     bool
	error       string
	initialized bool
	userID      string
	userName    string
	selectMode  bool
	agents      []AgentInfo
	selectedIdx int
}

type AgentInfo struct {
	ID     string
	Name   string
	Status string
}

// Message payloads
type MessageRequest struct {
	Text     string `json:"text"`
	SenderId string `json:"senderId"`
	RoomId   string `json:"roomId,omitempty"`
	Source   string `json:"source,omitempty"`
	EntityId string `json:"entityId,omitempty"`
	UserName string `json:"userName,omitempty"`
}

type MessageResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Message struct {
			Text string `json:"text"`
		} `json:"message"`
		MessageId string `json:"messageId"`
		Name      string `json:"name"`
		RoomId    string `json:"roomId"`
		Source    string `json:"source"`
	} `json:"data"`
}

type AgentsResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Agents []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"agents"`
	} `json:"data"`
}

func getCowsay(message string) (string, error) {
	return cowsay.Say(
		message,
		cowsay.Type("bender"),
	)
}

// Initial command to check for active agents
func checkAgents() tea.Cmd {
	return func() tea.Msg {
		resp, err := http.Get(apiBaseURL + "/agents")
		if err != nil {
			return errMsg{err: fmt.Errorf("failed to connect to Eliza API: %v", err)}
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return errMsg{err: fmt.Errorf("API returned status %d", resp.StatusCode)}
		}

		var agentsResp AgentsResponse
		if err := json.NewDecoder(resp.Body).Decode(&agentsResp); err != nil {
			return errMsg{err: fmt.Errorf("failed to decode response: %v", err)}
		}

		// Collect available agents
		agents := make([]AgentInfo, 0, len(agentsResp.Data.Agents))
		for _, agent := range agentsResp.Data.Agents {
			if agent.Status == "active" {
				agents = append(agents, AgentInfo{
					ID:     agent.ID,
					Name:   agent.Name,
					Status: agent.Status,
				})
			}
		}

		if len(agents) == 0 {
			return errMsg{err: fmt.Errorf("no active agents found")}
		}

		// If there's only one agent, select it automatically
		if len(agents) == 1 {
			return gotAgentMsg{id: agents[0].ID, name: agents[0].Name}
		}

		// Otherwise, let the user select from multiple agents
		return gotAgentsListMsg{agents: agents}
	}
}

// Command to send a message to the agent
func sendMessage(agentID, message, userID, userName string) tea.Cmd {
	return func() tea.Msg {
		msgReq := MessageRequest{
			Text:     message,
			SenderId: userID,
			Source:   "cowui",
			EntityId: userID,
			UserName: userName,
		}

		jsonData, err := json.Marshal(msgReq)
		if err != nil {
			return errMsg{err: fmt.Errorf("failed to encode message: %v", err)}
		}

		resp, err := http.Post(
			fmt.Sprintf("%s/agents/%s/message", apiBaseURL, agentID),
			"application/json",
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			return errMsg{err: fmt.Errorf("failed to send message: %v", err)}
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return errMsg{err: fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))}
		}

		var msgResp MessageResponse
		if err := json.NewDecoder(resp.Body).Decode(&msgResp); err != nil {
			return errMsg{err: fmt.Errorf("failed to decode response: %v", err)}
		}

		return gotResponseMsg{text: msgResp.Data.Message.Text}
	}
}

// Message types
type errMsg struct {
	err error
}

type gotAgentMsg struct {
	id   string
	name string
}

type gotAgentsListMsg struct {
	agents []AgentInfo
}

type gotResponseMsg struct {
	text string
}

// Initialize the model
func initialModel() Model {
	// Generate a persistent user ID
	userID := uuid.New().String()
	return Model{
		messages:    []string{},
		loading:     true,
		initialized: false,
		userID:      userID,
		userName:    "User",
		selectMode:  false,
		selectedIdx: 0,
		agents:      []AgentInfo{},
	}
}

func (m Model) Init() tea.Cmd {
	return checkAgents()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.selectMode {
			// Handle agent selection mode
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				return m, tea.Quit
			case tea.KeyUp:
				if m.selectedIdx > 0 {
					m.selectedIdx--
				}
				return m, nil
			case tea.KeyDown:
				if m.selectedIdx < len(m.agents)-1 {
					m.selectedIdx++
				}
				return m, nil
			case tea.KeyEnter:
				// Selected an agent
				selectedAgent := m.agents[m.selectedIdx]
				m.selectMode = false
				m.agentID = selectedAgent.ID
				m.initialized = true
				m.loading = false
				m.messages = append(m.messages, fmt.Sprintf("Connected to agent: %s", selectedAgent.Name))
				return m, nil
			}
			return m, nil
		}

		// Handle normal chat mode
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.userInput == "" || m.loading {
				return m, nil
			}

			// Add user message to history
			userMsg := "You: " + m.userInput
			m.messages = append(m.messages, userMsg)

			// Store the user input and clear it
			input := m.userInput
			m.userInput = ""
			m.loading = true

			// Send the message to the Eliza API
			return m, sendMessage(m.agentID, input, m.userID, m.userName)
		case tea.KeyBackspace:
			if len(m.userInput) > 0 {
				m.userInput = m.userInput[:len(m.userInput)-1]
			}
		case tea.KeySpace:
			m.userInput += " "
		case tea.KeyRunes:
			m.userInput += string(msg.Runes)
		}
	case gotAgentMsg:
		m.agentID = msg.id
		m.loading = false
		m.initialized = true
		m.messages = append(m.messages, fmt.Sprintf("Connected to agent: %s", msg.name))
		return m, nil
	case gotAgentsListMsg:
		m.agents = msg.agents
		m.selectMode = true
		m.loading = false
		return m, nil
	case gotResponseMsg:
		m.loading = false
		// Add agent response to history
		agentMsg := "Agent: " + msg.text
		m.messages = append(m.messages, agentMsg)
		return m, nil
	case errMsg:
		m.loading = false
		m.error = msg.err.Error()
		return m, nil
	}
	return m, nil
}

func (m Model) View() string {
	// If we're in agent selection mode, show the agent selection menu
	if m.selectMode {
		return m.agentSelectionView()
	}

	// Otherwise show the chat view
	return m.chatView()
}

func (m Model) agentSelectionView() string {
	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")).
		MarginBottom(1)

	listStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1).
		Width(60)

	// Title
	title := titleStyle.Render("Select an Agent")

	// Build agent list
	var list strings.Builder
	for i, agent := range m.agents {
		item := fmt.Sprintf("%s (%s)", agent.Name, agent.Status)

		if i == m.selectedIdx {
			// Highlight selected item
			item = "> " + item
			item = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("170")).Render(item)
		} else {
			item = "  " + item
		}

		list.WriteString(item + "\n")
	}

	instructions := "\nUse ↑/↓ to navigate, Enter to select, Ctrl+C to quit"

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		listStyle.Render(list.String()),
		instructions,
	)
}

func (m Model) chatView() string {
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

	// Show current message or status in the ASCII art
	var currentMessage string
	if m.error != "" {
		currentMessage = "Error: " + m.error
	} else if m.loading {
		currentMessage = "Loading..."
	} else if !m.initialized {
		currentMessage = "Connecting to Eliza API..."
	} else if len(m.messages) == 0 {
		currentMessage = "Ready to chat with Eliza OS agent!"
	} else {
		// Get the last agent message
		for i := len(m.messages) - 1; i >= 0; i-- {
			if strings.HasPrefix(m.messages[i], "Agent: ") {
				currentMessage = strings.TrimPrefix(m.messages[i], "Agent: ")
				break
			}
		}
		if currentMessage == "" {
			currentMessage = "Waiting for first response..."
		}
	}

	// Generate Bender ASCII art using the neocowsay library
	benderArt, err := getCowsay(currentMessage)
	if err != nil {
		benderArt = fmt.Sprintf("Error generating cowsay: %v", err)
	}

	// Input prompt
	promptPrefix := "Your message: "
	if m.loading {
		promptPrefix = "Waiting for response... "
	}
	prompt := promptPrefix + m.userInput
	if !m.loading {
		prompt += "_"
	}

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
