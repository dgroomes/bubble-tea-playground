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

func (f file) Title() string       { return f.filePath }
func (f file) FilterValue() string { return f.filePath }
func (f file) Description() string {
	if f.fetching {
		return "Fetching..."
	}
	if f.size == -1 {
		return "-"
	}

	return prettyPrintBytes(f.size)
}

type model struct {
	updateFn func()
	list     teaList.Model
	keys     key.Binding
}

func (m *model) Init() tea.Cmd {
	// We are opting out of Bubble Tea's desire to own the execution of our domain logic. I'm sympathetic to the Elm Architecture
	// (https://guide.elm-lang.org/architecture) but to me, I don't really understand why the framework wants control
	// over executing the "update function". I'd rather have my own code update my model as needed, and then signal to
	// the Bubble Tea TUI machinery that there is a "new generation of the model" and that it should re-render the view.
	// After all, the rendering model of Bubble Tea (and React) is to "blow away the old view and render a new one" (and
	// use a diffing algorithm to minimize expense). This is a very convenient programming paradigm.
	//
	// So long story short, we send a blank command to Bubble Tea in the "Init" function.
	return func() tea.Msg { return struct{}{} }
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	log.Printf("Update: %v\n", msg)

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:

		switch msg.String() {

		case "enter":
			selectedItem, ok := m.list.SelectedItem().(file)
			if ok {
				log.Println("[enter] [ok]")
				if selectedItem.fetching {
					log.Println("[enter] [ok] [fetching]")
					// todo tick?
				} else if selectedItem.size == -1 {
					currentIdx := m.list.Index()
					log.Printf("Fetching size for %s\n", selectedItem.filePath)
					selectedItem.fetching = true
					m.list.SetItem(currentIdx, selectedItem)
					go func() {
						FetchSize(&selectedItem)
						log.Printf("About to set item %d with value %v\n", currentIdx, selectedItem)
						// This is a race condition. If the item was removed from the list, then it's an error to add it
						// back.
						//
						// This reminds of React and lists. In React, you'll run into a "Warning: Each child in a list should have a unique “key” prop."
						// And this is well described by the docs:https://react.dev/learn/rendering-lists#keeping-list-items-in-order-with-key
						// I think I should just go with the "blow it all away and re-render" approach. That's like my
						// favorite advantage of this programming paradigm.
						m.list.SetItem(currentIdx, selectedItem)
						m.updateFn()
					}()
				}
			} else {
				log.Printf("Unable to cast selectedItem to 'file': %v . This is unexpected.\n", m.list.SelectedItem())
			}

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *model) View() string {
	return appStyle.Render(m.list.View())
}

func main() {
	// A combination of logs and the debugger have helped me debug issues in this program. Bubble Tea has an official
	// function to configure logging (https://github.com/charmbracelet/bubbletea/blob/3eb74e8d9dac487100b6d19ccc09b0c7820a6c7f/README.md?plain=1#L294)
	// and it has print/logging functions (https://github.com/charmbracelet/bubbletea/blob/3eb74e8d9dac487100b6d19ccc09b0c7820a6c7f/tea.go#L746)
	// but I couldn't get them to work. So, I'm just doing logging the direct way.
	f, err := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v\n", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer f.Close()
	log.SetOutput(f)

	listKeyMap := key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "summarize selected"))

	delegate := teaList.NewDefaultDelegate() // I haven't figured out what this delegate is for yet.
	list := teaList.New(make([]teaList.Item, 0), delegate, 0, 0)
	list.Title = "Git Project Files Summarizer"
	list.Styles.Title = titleStyle
	list.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeyMap,
		}
	}

	var programPtr = new(tea.Program)
	updateFn := func() {
		if programPtr == nil {
			log.Fatal("The update was called before the program pointer was set. This is an illegal state.")
		}

		// Signal to the Bubble Tea machinery that something in the model changed. This "do nothing" message will get
		// sent to the "Update" function and then soon afterward the "View" function will be called to re-render the
		// view.
		programPtr.Send(struct{}{})
	}

	// TODO I'd like to experiment with a 'domainModel' and a 'teaModel' to separate the domain logic from the TUI API.
	m := model{
		updateFn: updateFn,
		list:     list,
		keys:     listKeyMap,
	}

	p := tea.NewProgram(&m, tea.WithAltScreen())
	// Let's do some acrobatics to make sure our model has access to the Bubble Tea program value. This is not idiomatic
	// Bubble Tea code but this is what I prefer and like to experiment with.
	*programPtr = *p

	go func() {
		log.Println("Go routine is executing to find Git project files ...")
		files, err := listGitProjectFiles()
		if err != nil {
			log.Fatal(err)
		}

		// Artificially slow down the program to simulate a slow operation and get a visual effect in the TUI.
		time.Sleep(750 * time.Millisecond)

		log.Printf("Found %d files\n", len(files))

		if len(files) != 0 {
			// Is this the right way to turn a slice of 'file' into a slice of 'Item'?
			items := make([]teaList.Item, 0, len(files))
			for _, f := range files {
				items = append(items, f)
			}

			m.list.SetItems(items)
			m.list.Select(0)
		}

		updateFn()
	}()

	_, err = p.Run()
	if err != nil {
		log.Println("Error running program:", err)
		os.Exit(1)
	}
}
