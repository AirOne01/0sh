package zeroSsh

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
)

// Apps
const (
	AppMainMenu App = iota
	AppInfo     App = iota
)

type listedApp struct {
	title, desc string
	app         App
}

func (p listedApp) Title() string       { return p.title }
func (p listedApp) Desc() string        { return p.desc }
func (p listedApp) App() App            { return p.app }
func (p listedApp) FilterValue() string { return p.title }

type appDelegate struct{}

func (d appDelegate) Height() int                             { return 2 }
func (d appDelegate) Spacing() int                            { return 1 }
func (d appDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d appDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(listedApp)
	if !ok {
		return
	}

	str := fmt.Sprintf("%s\n%s", appTitleStyle.Copy().Render(i.title), appDescStyle.Render(i.desc))
	if index == m.Index() {
		str = fmt.Sprintf("%s\n%s", selectedAppTitleStyle.Render(i.title), appDescSelectedStyle.Render(i.desc))
	}

	fmt.Fprint(w, str)
}

func appMainMenu(pty ssh.Pty) list.Model {
	apps := []list.Item{
		listedApp{
			title: "Infos",
			desc:  "Show some informations about the current SSH session",
			app:   AppInfo,
		},
		listedApp{
			title: "Main menu",
			desc:  "The menu you are currently looking at :-)",
			app:   AppMainMenu,
		},
	}

	l := list.New(apps, appDelegate{}, pty.Window.Width, pty.Window.Height)
	l.Title = "Welcome to 0sh!"
	l.SetShowStatusBar(false)
	// l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return l
}
