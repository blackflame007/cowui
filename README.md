# CowUI

A stylish terminal-based chat client for interacting with AI agents through an ASCII art interface.

![CowUI Screenshot](docs/screenshot.png)

## Features

- 🤖 Connect to AI agents via a simple API
- 🖼️ Messages displayed through ASCII art characters (featuring Bender!)
- 👥 Select from multiple available agents
- 💬 Persistent chat history during session
- 🎨 Clean, modern terminal UI using Bubble Tea and Lipgloss

## Installation

```bash
go install github.com/blackflame007/cowui@latest
```

Or build from source:

```bash
git clone https://github.com/blackflame007/cowui.git
cd cowui
go build ./cmd/cowui
```

## Usage

Start the client with:

```bash
cowui
```

### Controls

- **Type** to compose a message
- **Enter** to send a message
- **Up/Down** to navigate agent selection (when multiple agents are available)
- **Ctrl+C** to exit

## Requirements

- Go 1.20 or higher
- A running agent server (default: http://localhost:3000/api)

## Configuration

The agent server URL can be configured by modifying the `apiBaseURL` constant in `main.go`.

## How It Works

CowUI connects to a local API server that hosts AI agents. Messages are sent to the selected agent, and responses are displayed in the terminal with ASCII art using a customized version of cowsay.

## Development

### Dependencies

- [Neo-cowsay](https://github.com/blackflame007/Neo-cowsay) - ASCII art generation
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal UI styling

### Building

```bash
go build -o cowui ./cmd/cowui
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

Distributed under the MIT License. See `LICENSE` for more information.

## Acknowledgements

- [Charm](https://charm.sh) for the amazing terminal UI libraries
- All contributors and supporters of the project
