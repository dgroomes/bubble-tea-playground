// A demo program that showcases the Bubble Tea TUI. Please see the README for more information.
//
// This is adapted from the official tutorials: https://github.com/charmbracelet/bubbletea/tree/master/tutorials
package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"strings"
)

func main() {

	initialModel := model{
		fileNameOptions: nil,
		cursor:          0,
		selected:        make(map[int]struct{}),
		executing:       false,
		fileSummaries:   nil,
	}

	p := tea.NewProgram(initialModel)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

// For when the files were listed.
type fileListingMsg struct {
	fileNames []string
}

// For when the file summarization is complete.
type summarizationMsg struct {
	fileSummaries []string
}

type model struct {
	fileNameOptions []string         // Files in the current working directory. The user chooses among these to summarize.
	cursor          int              // Which file item our cursor is pointing at
	selected        map[int]struct{} // Which file items are selected
	executing       bool             // Are we in the execution state?
	fileSummaries   []string         // The file fileSummaries.
}

// List the files in the current working directory.
func listFiles() tea.Msg {
	currentWorkingDir, err := os.Getwd()
	if err != nil {
		return err
	}

	dir, err := os.Open(currentWorkingDir)
	if err != nil {
		return err
	}

	defer func() {
		err := dir.Close()
		if err != nil {
			fmt.Printf("Failed to close the current working dir: %v\n", err)
		}
	}()

	fileNames, err := dir.Readdirnames(0)
	if err != nil {
		return err
	}

	return fileListingMsg{fileNames}
}

func (m model) Init() tea.Cmd {
	return listFiles
}

// Summarize the given files. Return summaries for each file.
func summarizeFiles(m model) ([]string, error) {
	fileNameOptions := m.fileNameOptions
	if fileNameOptions == nil {
		return nil, fmt.Errorf("invalid state: Attempted to summarize files, but the file listing was never initialized")
	}

	chosenFileNames := make([]string, 0)
	{
		for i, choice := range fileNameOptions {
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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case error:
		return m, tea.Quit

	case fileListingMsg:
		m.fileNameOptions = msg.fileNames

	case summarizationMsg:
		m.executing = false
		m.fileSummaries = msg.fileSummaries
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
					return summarizationMsg{summaries}
				}
			}

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.fileNameOptions)-1 {
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
	// Note that we're not returning a command. Is this not an error/unexpected branch?
	return m, nil
}

// The View method returns the exact text that will be written to the terminal emulator.
//
// This is much like ReactJS, where the "render" function returns an abstract representation of the HTML that will be
// rendered in the web page.
func (m model) View() string {
	if m.fileNameOptions == nil {
		return "Initializing..."
	}

	if m.executing {
		return "Executing..."
	}

	if m.fileSummaries != nil {
		return fmt.Sprintf("Complete! Below is a summary of the selected listFiles. It shows each file size, in bytes:\n%q\n", m.fileSummaries)
	}

	var textBuilder strings.Builder
	textBuilder.WriteString("What listFiles should we summarize?\n\n")

	for i, choice := range m.fileNameOptions {

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

		textBuilder.WriteString(fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice))
	}

	textBuilder.WriteString("\nPress 'e' to execute the file summarization. Press 'q' to quit.\n")

	// Send the text to the Bubble Tea framework. The framework will take care of rendering it to the terminal.
	return textBuilder.String()
}
