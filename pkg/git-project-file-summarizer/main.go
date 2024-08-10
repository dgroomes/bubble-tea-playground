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
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	teaList "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var appStyle = lipgloss.NewStyle().Padding(1, 2)

var titleStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FFFDF5")).
	Background(lipgloss.Color("#25A065")).
	Padding(0, 1)

type file struct {
	filePath string
	fetching bool
	size     int64 // -1 represents that the size has not yet been fetched.
}

func FetchSize(f *file) {
	fi, err := os.Stat(f.filePath)
	if err != nil {
		log.Fatal(err)
	}

	// Artificially slow down the program to simulate a slow operation.
	time.Sleep(750 * time.Millisecond)

	f.size = fi.Size()
	f.fetching = false
	log.Println("Fetched size for", f.filePath)
}

func (f file) Title() string { return f.filePath }

// FilterValue is this the value that search will use?
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

func prettyPrintBytes(bytes int64) string {
	const (
		KiB = 1024
		MiB = 1024 * KiB
		GiB = 1024 * MiB
	)

	switch {
	case bytes < KiB:
		return fmt.Sprintf("%d B", bytes)
	case bytes < MiB:
		return fmt.Sprintf("%.1f KiB", float64(bytes)/float64(KiB))
	case bytes < GiB:
		return fmt.Sprintf("%.1f MiB", float64(bytes)/float64(MiB))
	default:
		return fmt.Sprintf("%.1f GiB", float64(bytes)/float64(GiB))
	}
}

type model struct {
	program *tea.Program
	list    teaList.Model
	keys    key.Binding
}

func listGitProjectFiles() ([]file, error) {
	currentWorkingDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	repo, err := git.PlainOpen(currentWorkingDir)

	if err != nil {
		log.Fatal(err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		log.Fatal(err)
	}

	// I think this finds the exclude patterns in the .gitignore file in the Git repository in this directory.
	patterns, err := gitignore.ReadPatterns(worktree.Filesystem, nil)
	if err != nil {
		log.Fatal(err)
	}

	// I think 'worktree.Excludes' are the ignore patterns in maybe the home directory's .gitignore file. Not really
	//sure.
	patterns = append(patterns, worktree.Excludes...)

	m := gitignore.NewMatcher(patterns)

	var files []file

	err = filepath.WalkDir(".", func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		isDir := info.IsDir()

		// Split the path into its components. For example, if the path is "hello/README.md", the components will be
		// ["hello", "README.md"].
		pathComponents := strings.Split(filepath.Clean(path), string(filepath.Separator))

		ignored := m.Match(pathComponents, isDir)
		if err != nil {
			return err
		}

		if isDir {
			// If the directory is ignored then, we can speed up the file walking process by skipping the directory.
			if ignored {
				return filepath.SkipDir
			}

			if path == ".git" {
				return filepath.SkipDir
			}

			// We don't want to list directories, only files.
			return nil
		}

		if ignored {
			return nil
		}

		files = append(files, file{filePath: path, size: -1})
		return nil
	})

	return files, err
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
						m.program.Send(struct{}{})
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

	m := model{
		program: new(tea.Program),
		list:    list,
		keys:    listKeyMap,
	}
	p := tea.NewProgram(&m, tea.WithAltScreen())
	// Let's do some acrobatics to make sure our model has access to the Bubble Tea program value. This is not idiomatic
	// Bubble Tea code but this is what I prefer and like to experiment with.
	*m.program = *p

	go func() {
		log.Println("Go routine is executing to find Git project files ...")
		files, err := listGitProjectFiles()
		if err != nil {
			log.Fatal(err)
		}

		// Artificially slow down the program to simulate a slow operation.
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

		// We need to signal to the Bubble Tea machinery that something changed. This "do nothing" message will get
		// sent to the "Update" function and then soon afterward the "View" function will be called to re-render the view.
		p.Send(struct{}{})
	}()

	_, err = p.Run()
	if err != nil {
		log.Println("Error running program:", err)
		os.Exit(1)
	}
}
