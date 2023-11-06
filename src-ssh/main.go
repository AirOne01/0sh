package main

import (
	"context"
	"errors"
	"fmt"
	"io"
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

const (
	host = "localhost"
	port = 2022
)

var (
	titleStyle            = lipgloss.NewStyle().MarginLeft(2).PaddingLeft(1).PaddingRight(1).Background(lipgloss.ANSIColor(93)).Foreground(lipgloss.ANSIColor(255))
	paginationStyle       = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle             = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle         = lipgloss.NewStyle().Margin(1, 0, 2, 4)
	appTitleStyle         = lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1).Background(lipgloss.ANSIColor(240)).Foreground(lipgloss.ANSIColor(255))
	selectedAppTitleStyle = lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1).Background(lipgloss.ANSIColor(93)).Foreground(lipgloss.ANSIColor(255)).Underline(true)
	appDescStyle          = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(240))
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

type app struct {
	title, desc string
}

func (p app) Title() string       { return p.title }
func (p app) Desc() string        { return p.desc }
func (p app) FilterValue() string { return p.title }

type appDelegate struct{}

func (d appDelegate) Height() int                             { return 2 }
func (d appDelegate) Spacing() int                            { return 1 }
func (d appDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d appDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(app)
	if !ok {
		return
	}

	str := fmt.Sprintf("%s\n%s", appTitleStyle.Copy().Render(i.title), appDescStyle.Render(i.desc))
	if index == m.Index() {
		str = fmt.Sprintf("%s\n%s", selectedAppTitleStyle.Render(i.title), appDescStyle.Render(i.desc))
	}

	fmt.Fprint(w, str)
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

		apps := []list.Item{
			app{
				title: "Ethan",
				desc:  "ewilson@ymail.com",
			},
			app{
				title: "Joseph",
				desc:  "j_wright@live.com",
			},
			app{
				title: "Benjamin",
				desc:  "be@aol.com",
			},
			app{
				title: "Aaron",
				desc:  "aaron_jenkins@yahoo.com",
			},
			app{
				title: "Stephanie",
				desc:  "s_campbell@hotmail.com",
			},
			app{
				title: "Jose",
				desc:  "jjjackson50@outlook.com",
			},
			app{
				title: "Andrew",
				desc:  "awallen@outlook.com",
			},
			app{
				title: "Henry",
				desc:  "hareed@aol.com",
			},
			app{
				title: "Mary",
				desc:  "mary_clark@aol.com",
			},
			app{
				title: "Kyle",
				desc:  "kylebell@live.com",
			},
			app{
				title: "Laura",
				desc:  "ll@rocketmail.com",
			},
			app{
				title: "Alexander",
				desc:  "alexander_w_walker@gmail.com",
			},
			app{
				title: "Sophia",
				desc:  "sophiabailey84@outlook.com",
			},
			app{
				title: "Henry",
				desc:  "henry_e_taylor@aol.com",
			},
			app{
				title: "Sean",
				desc:  "sean_phillips44@live.com",
			},
			app{
				title: "Amelia",
				desc:  "ameliacook@live.com",
			},
		}

		l := list.New(apps, appDelegate{}, pty.Window.Width, pty.Window.Height)
		l.Title = "Welcome to 0sh!"
		l.SetShowStatusBar(false)
		// l.SetFilteringEnabled(false)
		l.Styles.Title = titleStyle
		l.Styles.PaginationStyle = paginationStyle
		l.Styles.HelpStyle = helpStyle

		m := model{
			list:     l,
			choice:   "",
			quitting: false,
			term:     pty.Term,
			height:   pty.Window.Height,
			width:    pty.Window.Width,
			time:     time.Now(),
			user:     session.User(),
			sys:      session.Subsystem(),
		}
		return newProg(m, tea.WithInput(session), tea.WithOutput(session), tea.WithAltScreen())
	}
	return bm.MiddlewareWithProgramHandler(teaHandler, termenv.ANSI256)
}

type model struct {
	list     list.Model
	choice   string
	quitting bool
	term     string
	width    int
	height   int
	time     time.Time
	user     string
	sys      string
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
		m.list.SetWidth(m.width)
		m.list.SetHeight(m.height)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			i, ok := m.list.SelectedItem().(app)
			if ok {
				m.choice = i.FilterValue()
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	// s := "Your term is %s\n"
	// s += "Your window size is x=%d, y=%d\n"
	// s += "Your username is %s\n"
	// s += "The current time is " + m.time.Format(time.RFC1123) + "\n"
	// s += "Press q or ctrl+c to quit"
	// s += docStyle.Render(m.list.View())
	// return fmt.Sprintf(s, m.term, m.width, m.height, m.user, m.sys)

	if m.choice != "" {
		return quitTextStyle.Render(fmt.Sprintf("%s? Sounds good to me.", m.choice))
	}
	if m.quitting {
		return quitTextStyle.Render("Not hungry? That’s cool.")
	}
	return "\n" + m.list.View()
}
