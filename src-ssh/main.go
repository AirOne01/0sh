package zeroSsh

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	lm "github.com/charmbracelet/wish/logging"
	"github.com/muesli/termenv"
)

type App int

const (
	host = "localhost"
	port = 2022
)

var (
	titleStyle            = lipgloss.NewStyle().MarginLeft(2).PaddingLeft(1).PaddingRight(1).Background(lipgloss.ANSIColor(99)).Foreground(lipgloss.ANSIColor(255))
	paginationStyle       = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle             = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	appTitleStyle         = lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1).Background(lipgloss.ANSIColor(240)).Foreground(lipgloss.ANSIColor(255))
	selectedAppTitleStyle = lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1).Background(lipgloss.ANSIColor(93)).Foreground(lipgloss.ANSIColor(255)).Bold(true).Underline(true)
	appDescStyle          = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(240))
	appDescSelectedStyle  = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(255))
)

func main() {
	var s, err = wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%d", host, port)),
		wish.WithMiddleware(
			bubbleTeaMiddleware(),
			lm.Middleware(),
		),
	)
	if err != nil {
		log.Error("could not start server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting SSH server", "host", host, "port", port)
	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("could not stop server", "error", err)
	}
}

func bubbleTeaMiddleware() wish.Middleware {
	newProg := func(model tea.Model, opts ...tea.ProgramOption) *tea.Program {
		program := tea.NewProgram(model, opts...)
		go func() {
			for {
				<-time.After(1 * time.Second)
				program.Send(timeMsg(time.Now()))
			}
		}()
		return program
	}
	teaHandler := func(session ssh.Session) *tea.Program {
		pty, _, active := session.Pty()
		if !active {
			wish.Fatalln(session, "no active terminal, skipping")
			return nil
		}

		m := model{
			appMainMenuList: appMainMenu(pty),
			appInfoTable:    appInfoTable(pty.Term, pty.Window.Width, pty.Window.Height, session.User(), session.Subsystem(), time.Now()),
			chosenApp:       AppMainMenu,
			term:            pty.Term,
			height:          pty.Window.Height,
			width:           pty.Window.Width,
			time:            time.Now(),
			user:            session.User(),
			sys:             session.Subsystem(),
		}
		return newProg(m, tea.WithInput(session), tea.WithOutput(session), tea.WithAltScreen())
	}
	return bm.MiddlewareWithProgramHandler(teaHandler, termenv.ANSI256)
}

type timeMsg time.Time

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case timeMsg:
		m.time = time.Time(msg)
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		// Main menu list
		m.appMainMenuList.SetWidth(m.width)
		m.appMainMenuList.SetHeight(m.height)
		// Info table
		m.appInfoTable.SetWidth(m.width)
		m.appInfoTable.SetHeight(m.height)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "backspace":
			m.chosenApp = AppMainMenu
		case "enter":
			i, ok := m.mappMainMenuListSelectedItem().(listedApp)
			if ok {
				m.chosenApp = i.App()
			}
		}
	}

	var cmd tea.Cmd
	// m.mappMainMenuList cmd = m.mappMainMenuListUpdate(msg)
	return m, cmd
}

func (m model) View() string {
	switch m.chosenApp {
	case AppInfo:
		return m.appInfoTable.View()
	default:
		return m.mappMainMenuListView()
	}
}
