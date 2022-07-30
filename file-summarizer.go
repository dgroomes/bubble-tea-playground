package main

import (
	"fmt"
	"os"
)
import tea "github.com/charmbracelet/bubbletea"

// A demo program that showcases the Bubble Tea TUI. Please see the README for more information.
//
// This is adapted from the official tutorials: https://github.com/charmbracelet/bubbletea/tree/master/tutorials
func main() {

	initialModel := model{
		choices:   nil,
		cursor:    0,
		selected:  make(map[int]struct{}),
		executing: false,
		summaries: nil,
	}

	p := tea.NewProgram(initialModel)

	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

type model struct {
	choices   []string         // Files in the current working directory. The user chooses which to summarize.
	cursor    int              // Which file item our cursor is pointing at
	selected  map[int]struct{} // While file items are selected
	executing bool             // Are we in the execution state?
	summaries []string         // The file summaries.
}

func files() tea.Msg {
	currentWorkingDir, _ := os.Getwd()

	file, _ := os.Open(currentWorkingDir)

	names, _ := file.Readdirnames(0)

	return names
}

func (m model) Init() tea.Cmd {
	return files
}

// Summarize the given files. Return summaries for each file.
func summarizeFiles(m model) (summaries, error) {

	chosenFileNames := make([]string, 0)
	{
		for i, choice := range m.choices {
			_, exists := m.selected[i]
			if exists {
				chosenFileNames = append(chosenFileNames, choice)
			}
		}
	}

	fileSummaries := make([]string, len(chosenFileNames))
	{
		for idx, name := range chosenFileNames {
			fi, err := os.Stat(name)
			if err != nil {
				return nil, err
			}
			// get the size
			size := fi.Size()
			fileSummaries[idx] = fmt.Sprintf("%s: %d", name, size)
		}
	}

	return fileSummaries, nil
}

type summaries []string

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case error:
		return m, tea.Quit

	case []string:
		// The initialization command has completed. We have the list of file names.
		// This is bad Go program design. Really, I should use a specific type... but I'm just learning here.
		m.choices = msg

	case summaries:
		m.executing = false
		m.summaries = msg
		return m, tea.Quit

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		case "e":
			m.executing = true
			return m, func() tea.Msg {
				summaries, err := summarizeFiles(m)
				if err != nil {
					return err
				} else {
					return summaries
				}
			}

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		// The "enter" key and the space bar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

// The View method returns the exact text that will be written to the terminal emulator.
//
// This is much like ReactJS, where the "render" function returns an abstract representation of the HTML that will be
// rendered in the web page.
func (m model) View() string {
	if m.choices == nil {
		return "Initializing..."
	}

	if m.executing {
		return "Executing..."
	}

	if m.summaries != nil {
		return fmt.Sprintf("Complete! Below is a summary of the selected files. It shows each file size, in bytes:\n%q\n", m.summaries)
	}

	text := "What files should we summarize?\n\n"

	for i, choice := range m.choices {

		// Tangential note: Go always automatically initializes variables. While in Java, I can rely on Java's compiler
		// to tell me something to the effect of "Hey you didn't initialize this variable, you should have an else branch".
		// We don't have this feature in Go. Instead, it might be smart to be in the habit of always initializing a
		// variable in the same statement that declares it, or to rely on a linter to tell you that you missed a case.
		var cursor string
		if m.cursor == i {
			cursor = ">"
		} else {
			cursor = " "
		}

		var checked string
		{
			_, exists := m.selected[i]
			if exists {
				checked = "x"
			} else {
				checked = " "
			}
		}

		text += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}

	text += "\nPress 'e' to execute the file summarization. Press 'q' to quit.\n"

	// Send the text to the Bubble Tea framework. The framework will take care of rendering it to the terminal.
	return text
}
