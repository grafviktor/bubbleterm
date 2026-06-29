package manager

import "charm.land/lipgloss/v2"

var paneStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("240"))

var focusedStyle = paneStyle.BorderForeground(lipgloss.Color("205"))
