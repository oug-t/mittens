package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"mittens/cmd/internal/agent"
	"mittens/cmd/internal/engine"
)

var (
	nordBlue   = lipgloss.Color("#81A1C1")
	nordGrey   = lipgloss.Color("#4C566A")
	nordActive = lipgloss.Color("#88C0D0")

	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(nordBlue).MarginBottom(1)
	boxStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(nordGrey).Padding(1, 2)
	helpStyle  = lipgloss.NewStyle().Foreground(nordGrey).MarginTop(1)
)

type Mode int

const (
	NormalMode Mode = iota
	InsertMode
)

type AgentResponseMsg struct {
	Output string
	Err    error
}

type Model struct {
	vm        *engine.VM
	vmStatus  string
	logs      []string
	mode      Mode
	textInput textinput.Model
	viewport  viewport.Model
	spinner   spinner.Model
	isWorking bool
}

func InitialModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Press 'i' to enter Insert Mode..."

	vp := viewport.New(50, 10)
	vp.SetContent("Waiting for agent output...")

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(nordActive)

	return Model{
		vmStatus:  "OFFLINE",
		logs:      []string{"System ready."},
		mode:      NormalMode,
		textInput: ti,
		viewport:  vp,
		spinner:   s,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
		spCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case NormalMode:
			switch msg.String() {
			case "q", "ctrl+c":
				if m.vm != nil {
					m.vm.Stop()
				}
				return m, tea.Quit
			case "s":
				m.vmStatus = "BOOTING"
				m.viewport.SetContent("Booting microVM...")
				sockPath := fmt.Sprintf("/tmp/fc-%d.sock", time.Now().UnixNano())
				vm, err := engine.NewVM(sockPath, "vmlinux", "rootfs.ext4")
				if err != nil {
					m.vmStatus = "ERROR"
					m.viewport.SetContent(fmt.Sprintf("Failed to build VM:\n%v", err))
				} else if err := vm.Start(); err != nil {
					m.vmStatus = "ERROR"
					m.viewport.SetContent(fmt.Sprintf("Failed to start VM:\n%v", err))
				} else {
					m.vm = vm
					m.vmStatus = "RUNNING"
					m.viewport.SetContent("VM Running! Press 'i' to enter prompt.")
				}
			case "k":
				if m.vm != nil {
					m.vm.Stop()
					m.vm = nil
				}
				m.vmStatus = "OFFLINE"
			case "i":
				m.mode = InsertMode
				return m, m.textInput.Focus()
			}
		case InsertMode:
			switch msg.String() {
			case "esc":
				m.mode = NormalMode
				m.textInput.Blur()
			case "enter":
				val := m.textInput.Value()
				if m.vmStatus == "RUNNING" {
					m.isWorking = true
					m.textInput.Reset()
					return m, func() tea.Msg {
						out, err := agent.SendCommand(val)
						return AgentResponseMsg{Output: out, Err: err}
					}
				}
			}
		}

	case AgentResponseMsg:
		m.isWorking = false
		if msg.Err != nil {
			m.viewport.SetContent("Error: " + msg.Err.Error())
		} else {
			m.viewport.SetContent(msg.Output)
		}

	case spinner.TickMsg:
		m.spinner, spCmd = m.spinner.Update(msg)
	}

	if m.mode == InsertMode {
		m.textInput, tiCmd = m.textInput.Update(msg)
	}

	m.viewport, vpCmd = m.viewport.Update(msg)
	return m, tea.Batch(tiCmd, vpCmd, spCmd)
}

func (m Model) View() string {
	modeStr := " NORMAL "
	if m.mode == InsertMode {
		modeStr = " INSERT "
	}

	header := lipgloss.JoinHorizontal(lipgloss.Center,
		lipgloss.NewStyle().Background(nordActive).Foreground(lipgloss.Color("#2E3440")).Render(modeStr),
		titleStyle.PaddingLeft(1).Render("Mittens Sandbox"),
	)

	var content string
	if m.isWorking {
		content = fmt.Sprintf("\n\n%s Agent is thinking...\n", m.spinner.View())
	} else {
		content = m.viewport.View()
	}

	leftBox := boxStyle.Width(50).Height(14).Render(content)
	stats := fmt.Sprintf("AGENT POOL\n\nStatus: %s\nCID: 3\nNet: tap0\nArch: x86_64", m.vmStatus)
	rightBox := boxStyle.Width(25).Height(14).Render(stats)

	mainBody := lipgloss.JoinHorizontal(lipgloss.Top, leftBox, rightBox)

	return fmt.Sprintf("\n%s\n\n%s\n\n%s\n\n%s\n",
		header,
		mainBody,
		m.textInput.View(),
		helpStyle.Render("s: start | k: kill | i: insert | esc: normal | q: quit"))
}
