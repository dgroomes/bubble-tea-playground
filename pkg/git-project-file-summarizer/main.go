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
	updateFn        func()
	domainListModel *[]*file // I'm going for "reference type vibes". Not idiomatic Go, but it's how I know how to program.
	teaListModel    *teaList.Model
	keys            key.Binding
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
	log.Printf("[Update] msg=%v teaListModelLength=%d\n", msg, len(m.teaListModel.Items()))

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.teaListModel.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:

		switch msg.String() {

		case "enter":
			selectedItem, ok := m.teaListModel.SelectedItem().(*file)
			if ok {
				log.Println("[enter] [ok]")
				if selectedItem.fetching {
					log.Println("[enter] [ok] [fetching]")
					// todo tick?
				} else if selectedItem.size == -1 {
					go func() {
						log.Printf("Fetching size for %s\n", selectedItem.filePath)
						selectedItem.fetching = true
						m.updateFn()
						FetchSize(selectedItem)
						m.updateFn()
					}()
				}
			} else {
				log.Fatalf("Unable to cast selectedItem to 'file': %v . This is unexpected.\n", m.teaListModel.SelectedItem())
			}

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	newListModel, cmd := m.teaListModel.Update(msg)
	*m.teaListModel = newListModel
	log.Printf("[Update] newListModel=%d\n", len(newListModel.Items()))
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *model) View() string {
	return appStyle.Render(m.teaListModel.View())
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
	teaListModel := teaList.New(make([]teaList.Item, 0), delegate, 0, 0)
	teaListModel.Title = "Git Project Files Summarizer"
	teaListModel.Styles.Title = titleStyle
	teaListModel.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeyMap,
		}
	}

	var programPtr = new(tea.Program)
	var modelPtr = new(model)

	// The updateFn is a bridge between my own domain model and the Bubble Tea machinery. It's similar in spirit to
	// Bubble Tea's "Update" function, but it isn't message driven. Or you might say, it isn't inspired by the Elm
	// Architecture. I'm not passionate about the design I've come up with, but just experimenting.
	updateFn := func() {
		log.Println("[updateFn]")
		if programPtr == nil {
			log.Fatal("[updateFn] updateFn was called before the program pointer was set. This is an illegal state.")
		}
		if modelPtr == nil {
			log.Fatal("[updateFn] updateFn was called before the model pointer was set. This is an illegal state.")
		}

		files := *modelPtr.domainListModel

		if len(files) != 0 {
			log.Println("[updateFn] 'files' is not empty. Adapting the domain model into the TUI model.")

			// Here specifically, we're bridging the domain models' 'set of files' to the model of the 'list' TUI
			// component in the Bubbles component library. I'm going with a React-style "blow it all away and re-render"
			// approach. Although here, we're not at the render stage/view stage we're actually at an earlier
			// "sync the models" stage. This is a wasteful operation, but I like the programming model.

			// First, take note of the selected item in the list. We'll try to preserve this selection.
			var selectedItem *file
			if modelPtr.teaListModel.SelectedItem() == nil {
				log.Println("[updateFn] The Bubbles list has no selected item. This is normal at startup time.")
			} else {
				x, ok := modelPtr.teaListModel.SelectedItem().(*file)
				if !ok {
					log.Fatal("[updateFn] The 'SelectedItem' in the Bubbles list is not a pointer to 'file'. This is unexpected. (But couldn't it be nil?)")
				}
				selectedItem = x
			}

			desiredIndex := -1
			items := make([]teaList.Item, 0, len(files))
			for i, f := range files {
				items = append(items, f)
				// This reminds of React and lists. In React, you'll run into a "Warning: Each child in a list should have a unique “key” prop."
				// And this is well described by the docs:https://react.dev/learn/rendering-lists#keeping-list-items-in-order-with-key
				//
				// I'm not using a key. Instead of comparing based on pointer equality, which is dubious because there's
				// no guarantee that any given address will point the same data in the future.
				if selectedItem == f {
					desiredIndex = i
				}
			}
			modelPtr.teaListModel.SetItems(items)

			// Re-select the original item. We have to do this because there is a chance that the list has shifted
			// (elements added or removed) during this update operation (well there is no chance because I didn't
			// program it that way, but I'm playing pretend)
			if desiredIndex >= 0 { // Not totally sure I need this check
				log.Printf("[updateFn] Reselecting item at index %d\n", desiredIndex)
				modelPtr.teaListModel.Select(desiredIndex)
			} else {
				log.Println("[updateFn] There was no existing item selection. I think this is normal at startup time.")
			}
		} else {
			log.Println("[updateFn] 'files' is empty.")
		}

		// Signal to the Bubble Tea machinery that something in the model changed. This "do nothing" message will get
		// sent to the "Update" function and then soon afterward the "View" function will be called to re-render the
		// view.
		programPtr.Send(struct{}{})
	}

	// TODO I'd like to experiment with a 'domainModel' and a 'teaModel' to separate the domain logic from the TUI API.
	m := model{
		updateFn:        updateFn,
		teaListModel:    &teaListModel,
		domainListModel: &[]*file{},
		keys:            listKeyMap,
	}

	p := tea.NewProgram(&m, tea.WithAltScreen())
	// Let's do some acrobatics to make sure our model has access to the Bubble Tea program value. This is not idiomatic
	// Bubble Tea code but this is what I prefer and like to experiment with.
	*programPtr = *p
	*modelPtr = m

	go func() {
		log.Println("Go routine is executing to find Git project files ...")
		files, err := listGitProjectFiles()
		if err != nil {
			log.Fatal(err)
		}

		// Artificially slow down the program to simulate a slow operation and get a visual effect in the TUI.
		time.Sleep(750 * time.Millisecond)

		log.Printf("Found %d files\n", len(files))

		*modelPtr.domainListModel = files
		updateFn()
	}()

	_, err = p.Run()
	if err != nil {
		log.Println("Error running program:", err)
		os.Exit(1)
	}
}
