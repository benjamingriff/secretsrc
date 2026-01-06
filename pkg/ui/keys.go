package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all key bindings for the application
type KeyMap struct {
	Up           key.Binding
	Down         key.Binding
	Left         key.Binding
	Right        key.Binding
	Select       key.Binding
	Back         key.Binding
	ViewValue    key.Binding
	CopyPlain    key.Binding
	CopyJSON     key.Binding
	Refresh      key.Binding
	Profile      key.Binding
	Region       key.Binding
	NextPage     key.Binding
	PrevPage     key.Binding
	Filter       key.Binding
	GridNextPage key.Binding
	GridPrevPage key.Binding
	Help         key.Binding
	Quit         key.Binding
}

// DefaultKeyMap returns the default key bindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "move left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "move right"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc", "q"),
			key.WithHelp("esc/q", "back/quit"),
		),
		ViewValue: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "view secret"),
		),
		CopyPlain: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "copy plain"),
		),
		CopyJSON: key.NewBinding(
			key.WithKeys("j"),
			key.WithHelp("j", "copy json"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Profile: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "profile"),
		),
		Region: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "region"),
		),
		NextPage: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "next AWS page"),
		),
		PrevPage: key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "prev AWS page"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		GridNextPage: key.NewBinding(
			key.WithKeys(" ", "pgdown"),
			key.WithHelp("space/pgdn", "next screen"),
		),
		GridPrevPage: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("pgup", "prev screen"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
}
