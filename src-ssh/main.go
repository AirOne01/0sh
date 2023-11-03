package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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
		model := DataModel{
			term:   pty.Term,
			height: pty.Window.Height,
			width:  pty.Window.Width,
			time:   time.Now(),
			user:   session.User(),
			sys:    session.Subsystem(),
		}
		return newProg(model, tea.WithInput(session), tea.WithOutput(session), tea.WithAltScreen())
	}
	return bm.MiddlewareWithProgramHandler(teaHandler, termenv.ANSI256)
}

type DataModel struct {
	term   string
	width  int
	height int
	time   time.Time
	user   string
	sys    string
}

type timeMsg time.Time

func (model DataModel) Init() tea.Cmd {
	return nil
}

func (model DataModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case timeMsg:
		model.time = time.Time(msg)
	case tea.WindowSizeMsg:
		model.height = msg.Height
		model.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return model, tea.Quit
		}
	}
	return model, nil
}

func (model DataModel) View() string {
	s := "Your term is %s\n"
	s += "Your window size is x=%d, y=%d\n"
	s += "Your username is %s\n"
	s += "The current time is " + model.time.Format(time.RFC1123) + "\n"
	s += "Press q or ctrl+c to quit"
	return fmt.Sprintf(s, model.term, model.width, model.height, model.user, model.sys)
}
