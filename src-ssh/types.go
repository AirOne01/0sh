package zeroSsh

import (
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss/table"
)

type model struct {
	appMainMenuList list.Model  // AppMainMenu
	appInfoTable    table.Table // AppInfo
	chosenApp       App
	term            string
	width           int
	height          int
	time            time.Time
	user            string
	sys             string
}
