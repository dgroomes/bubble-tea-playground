// A demo program that uses TUI components from the "Bubbles" component library (https://github.com/charmbracelet/bubbles)
// like "list" and "spinner" (not yet implemented). This is an intermediate Bubble Tea example program that builds above the basic example
// program in "file-summarizer.go". This program is also adapted from the "Fancy List" official example
// program: https://github.com/charmbracelet/bubbletea/tree/master/examples/list-fancy
//
// This program is a TUI (Text User Interface) and it presents a list of all files in the Git project. The list is
// interactive. It can be real-time filtered by typing a glob search, and you can select files to "summarize". The
// summarization will just fetch the file size. It's a toy example, but it should give you a good idea of how to design
// a TUI program using Bubble Tea.
package main

import (
	"github.com/charmbracelet/bubbles/key"
	teaList "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"log"
	"os"
	"time"
)

var appStyle = lipgloss.NewStyle().Padding(1, 2)

var titleStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FFFDF5")).
	Background(lipgloss.Color("#25A065")).
	Padding(0, 1)

func (f File) Title() string       { return f.filePath }
func (f File) FilterValue() string { return f.filePath }
func (f File) Description() string {
	if f.fetching {
		return "Fetching..."
	}
	if f.size == -1 {
		return "-"
	}

	return prettyPrintBytes(f.size)
}

type model struct {
	teaListModel teaList.Model
	keys         key.Binding
}

type foundFiles []File

func (m model) Init() tea.Cmd {
	return func() tea.Msg {
		log.Println("Go routine is executing to find Git project files ...")
		files, err := listGitProjectFiles()
		if err != nil {
			log.Fatal(err)
		}

		// Artificially slow down the program to simulate a slow operation and get a visual effect in the TUI.
		time.Sleep(750 * time.Millisecond)

		log.Printf("Found %d files\n", len(files))

		return foundFiles(files)
	}
}

type afterFetch File

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	log.Printf("[Update] msg=%v\n", msg)

	switch msg := msg.(type) {

	case foundFiles:
		items := make([]teaList.Item, 0, len(msg))
		for _, f := range msg {
			items = append(items, f)
		}
		m.teaListModel.SetItems(items)
		return m, nil

	case afterFetch:
		for i, item := range m.teaListModel.Items() {
			f, ok := item.(File)
			if !ok {
				log.Fatalf("The 'Item' in the Bubbles list is not a 'File'. This is unexpected.\n")
			}

			if f.filePath == msg.filePath {
				m.teaListModel.SetItem(i, File(msg))
				break
			}
		}

	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.teaListModel.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:

		switch msg.String() {

		case "enter":
			if len(m.teaListModel.Items()) == 0 {
				log.Println("No files to summarize.")
				return m, nil
			}

			selectedItem, ok := m.teaListModel.SelectedItem().(File)
			if !ok {
				log.Fatalf("The 'SelectedItem' in the Bubbles list is not a 'File'. This is unexpected.\n")
			}

			if selectedItem.size != -1 || selectedItem.fetching {
				log.Println("Already fetched or fetching. Not fetching again.")
				return m, nil
			}

			selectedItem.fetching = true
			idx := m.teaListModel.Index()
			m.teaListModel.SetItem(idx, selectedItem)
			return m, func() tea.Msg {
				return afterFetch(selectedItem.FetchSize())
			}

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	newListModel, cmd := m.teaListModel.Update(msg)
	m.teaListModel = newListModel
	log.Printf("[Update] newListModel=%d\n", len(newListModel.Items()))
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	return appStyle.Render(m.teaListModel.View())
}

func main() {
	// A combination of logs and the debugger have helped me debug issues in this program. Bubble Tea has an official
	// function to configure logging (https://github.com/charmbracelet/bubbletea/blob/3eb74e8d9dac487100b6d19ccc09b0c7820a6c7f/README.md?plain=1#L294)
	// and it has print/logging functions (https://github.com/charmbracelet/bubbletea/blob/3eb74e8d9dac487100b6d19ccc09b0c7820a6c7f/tea.go#L746)
	// but I couldn't get them to work. So, I'm just doing logging the direct way.
	f, err := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening File: %v\n", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer f.Close()
	log.SetOutput(f)

	listKeyMap := key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "summarize selected"))

	delegate := teaList.NewDefaultDelegate() // I haven't figured out what this delegate is for yet.
	teaListModel := teaList.New(make([]teaList.Item, 0), delegate, 0, 0)
	teaListModel.Title = "Git Project Files Summarizer"
	teaListModel.Styles.Title = titleStyle
	teaListModel.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeyMap,
		}
	}

	m := model{
		teaListModel: teaListModel,
		keys:         listKeyMap,
	}

	p := tea.NewProgram(&m, tea.WithAltScreen())

	_, err = p.Run()
	if err != nil {
		log.Println("Error running program:", err)
		os.Exit(1)
	}
}
