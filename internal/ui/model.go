package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shirou/gopsutil/v4/process"
	procinfo "github.com/uncaughtx/portcheck/internal/process"
)

// State represents the current UI state
type State int

const (
	StateMenu State = iota
	StateInfo
	StateConfirmKill
	StateKilling
	StateDone
)

// Model is the bubbletea model for the interactive UI
type Model struct {
	port      int
	process   *procinfo.Info
	state     State
	err       error
	width     int
	height    int
	confirmed bool
	killed    bool
}

// NewModel creates a new UI model
func NewModel(port int, proc *procinfo.Info) Model {
	return Model{
		port:    port,
		process: proc,
		state:   StateMenu,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case StateMenu:
			return m.handleMenuKey(msg)
		case StateInfo:
			return m.handleInfoKey(msg)
		case StateConfirmKill:
			return m.handleConfirmKey(msg)
		case StateDone:
			return m.handleDoneKey(msg)
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case killResultMsg:
		m.state = StateDone
		m.err = msg.err
		m.killed = msg.err == nil
	}
	return m, nil
}

func (m Model) handleMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "k":
		m.state = StateConfirmKill
	case "i":
		m.state = StateInfo
	case "q", "esc", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleInfoKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "b", "esc":
		m.state = StateMenu
	case "k":
		m.state = StateConfirmKill
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		m.state = StateKilling
		return m, m.killProcess
	case "n", "N", "esc":
		m.state = StateMenu
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleDoneKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "ctrl+c", "enter":
		return m, tea.Quit
	}
	return m, nil
}

type killResultMsg struct{ err error }

func (m Model) killProcess() tea.Msg {
	proc, err := process.NewProcess(m.process.PID)
	if err != nil {
		return killResultMsg{err}
	}
	err = proc.Terminate()
	return killResultMsg{err}
}

// View renders the UI
func (m Model) View() string {
	switch m.state {
	case StateMenu:
		return m.menuView()
	case StateInfo:
		return m.infoView()
	case StateConfirmKill:
		return m.confirmView()
	case StateKilling:
		return m.killingView()
	case StateDone:
		return m.doneView()
	}
	return ""
}

func (m Model) menuView() string {
	var b strings.Builder

	// Header
	b.WriteString(BoxStyle.Render(fmt.Sprintf("  Port %d is in use  ", m.port)))
	b.WriteString("\n\n")

	// Process info summary
	b.WriteString(fmt.Sprintf("  %s %s\n", LabelStyle.Render("Process:"), ValueStyle.Render(m.process.Name)))
	b.WriteString(fmt.Sprintf("  %s %s\n", LabelStyle.Render("PID:"), HighlightStyle.Render(fmt.Sprintf("%d", m.process.PID))))
	b.WriteString(fmt.Sprintf("  %s %s\n", LabelStyle.Render("Command:"), ValueStyle.Render(Truncate(m.process.Cmdline, 50))))
	b.WriteString(fmt.Sprintf("  %s %s\n", LabelStyle.Render("Running:"), ValueStyle.Render(m.process.Uptime())))

	if m.process.Cwd != "" {
		b.WriteString(fmt.Sprintf("  %s %s\n", LabelStyle.Render("Directory:"), ValueStyle.Render(Truncate(m.process.Cwd, 50))))
	}

	if m.process.ParentName != "" {
		b.WriteString(fmt.Sprintf("  %s %s (PID %d)\n",
			LabelStyle.Render("Parent:"),
			ValueStyle.Render(m.process.ParentName),
			m.process.ParentPID))
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render(fmt.Sprintf("  %s kill  %s info  %s quit",
		HelpKeyStyle.Render("[k]"),
		HelpKeyStyle.Render("[i]"),
		HelpKeyStyle.Render("[q]"))))

	return b.String()
}

func (m Model) infoView() string {
	var b strings.Builder

	// Header
	b.WriteString(BoxStyle.Render(fmt.Sprintf("  Process Details — PID %d  ", m.process.PID)))
	b.WriteString("\n\n")

	// Full process info
	b.WriteString(fmt.Sprintf("  %s %s\n", LabelStyle.Render("Name:"), ValueStyle.Render(m.process.Name)))
	b.WriteString(fmt.Sprintf("  %s %s\n", LabelStyle.Render("PID:"), HighlightStyle.Render(fmt.Sprintf("%d", m.process.PID))))
	b.WriteString(fmt.Sprintf("  %s %s\n", LabelStyle.Render("User:"), ValueStyle.Render(m.process.User)))
	b.WriteString(fmt.Sprintf("  %s %s\n", LabelStyle.Render("Running:"), ValueStyle.Render(m.process.Uptime())))

	if m.process.ProjectName != "" {
		b.WriteString(fmt.Sprintf("  %s %s\n", LabelStyle.Render("Project:"), HighlightStyle.Render(m.process.ProjectName)))
	}

	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  %s\n", LabelStyle.Render("Command:")))
	b.WriteString(fmt.Sprintf("  %s\n", ValueStyle.Render(m.process.Cmdline)))

	if m.process.Cwd != "" {
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  %s\n", LabelStyle.Render("Working Directory:")))
		b.WriteString(fmt.Sprintf("  %s\n", ValueStyle.Render(m.process.Cwd)))
	}

	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  %s %.1f MB\n", LabelStyle.Render("Memory:"), m.process.MemoryMB))
	b.WriteString(fmt.Sprintf("  %s %.1f%%\n", LabelStyle.Render("CPU:"), m.process.CPUPercent))
	b.WriteString(fmt.Sprintf("  %s %d\n", LabelStyle.Render("Children:"), m.process.ChildCount))

	if m.process.ParentName != "" {
		b.WriteString(fmt.Sprintf("  %s %s (PID %d)\n",
			LabelStyle.Render("Parent:"),
			ValueStyle.Render(m.process.ParentName),
			m.process.ParentPID))
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render(fmt.Sprintf("  %s back  %s kill  %s quit",
		HelpKeyStyle.Render("[b]"),
		HelpKeyStyle.Render("[k]"),
		HelpKeyStyle.Render("[q]"))))

	return b.String()
}

func (m Model) confirmView() string {
	var b strings.Builder

	b.WriteString(BoxStyle.Render(fmt.Sprintf("  Terminate Process?  ")))
	b.WriteString("\n\n")

	b.WriteString(DangerStyle.Render("You are about to kill:"))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("  %s %s\n", LabelStyle.Render("Process:"), ValueStyle.Render(m.process.Name)))
	b.WriteString(fmt.Sprintf("  %s %s\n", LabelStyle.Render("PID:"), HighlightStyle.Render(fmt.Sprintf("%d", m.process.PID))))

	if m.process.ChildCount > 0 {
		b.WriteString("\n")
		b.WriteString(WarningStyle.Render(fmt.Sprintf("  ⚠ This process has %d child process(es)", m.process.ChildCount)))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Are you sure? %s yes  %s no",
		HelpKeyStyle.Render("[y]"),
		HelpKeyStyle.Render("[n]")))

	return b.String()
}

func (m Model) killingView() string {
	return WarningStyle.Render("\n  Terminating process...")
}

func (m Model) doneView() string {
	var b strings.Builder

	if m.killed {
		b.WriteString(SuccessStyle.Render(fmt.Sprintf("\n  ✓ Process %d terminated successfully\n", m.process.PID)))
		b.WriteString(fmt.Sprintf("\n  Port %d is now available", m.port))
	} else if m.err != nil {
		b.WriteString(ErrorStyle.Render(fmt.Sprintf("\n  ✗ Failed to terminate process: %v", m.err)))
	}

	b.WriteString("\n\n")
	b.WriteString(HelpStyle.Render("  Press any key to exit"))

	return b.String()
}
