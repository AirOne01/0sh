package zeroSsh

import (
	"strconv"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func appInfoTable(term string, width int, height int, user string, sys string, currentTime time.Time) *table.Table {
	rows := [][]string{
		{"Terminal type", term},
		{"Window width", strconv.Itoa(width)},
		{"Window height", strconv.Itoa(height)},
		{"Username", user},
		{"Subsystem", sys},
		{"Current time", currentTime.Format(time.RFC1123)},
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(99))).
		Headers("Key", "Value").
		Rows(rows...)

	return t
}
