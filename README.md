# bubble-tea-playground

📚 Learning and exploring the Go-based TUI framework: Bubble Tea.

> The fun, functional and stateful way to build terminal apps.
>
> -- <cite>https://github.com/charmbracelet/bubbletea</cite>


## Description

[Charm](https://github.com/charmbracelet) is building perhaps the best overall CLI tools in existence. They build a
collection of open-source projects that have a particular focus on making CLI experiences sophisticated, stylish,
and smooth around the edges.

I'd like to start by learning how to use Charm's TUI (Text User Interface) framework: Bubble Tea.


## Instructions

Follow these instructions to build and run some Bubble Tea example programs.

1. Build and run the file summarizer program:
   * ```shell
     go run bubble_tea_playground/pkg/file-summarizer
     ```
   * Altogether it will look something like the following.
   * ```text
     go run .

     What files should we summarize?

     > [ ] go.mod
       [ ] go.sum
       [ ] README.md
       [ ] .gitignore
       [ ] file-summarizer.go
       [ ] .idea

     Press 'e' to execute the file summarization. Press 'q' to quit.
     ```
   * ```text
     What files should we summarize?

       [ ] go.mod
       [ ] go.sum
       [x] README.md
       [ ] .gitignore
       [ ] file-summarizer.go
     > [x] .idea

     Press 'e' to execute the file summarization. Press 'q' to quit.
     ```
   * ```text
     Complete! Below is a summary of the selected files. It shows each file size, in bytes:
     ["README.md: 1630" ".idea: 192"]
     ```
2. Now, let's implement similar functionality but in a fancy list TUI component.
    * ```shell
      go run bubble_tea_playground/pkg/git-project-file-summarizer
      ```


## Wish List

General clean-ups, TODOs and things I wish to implement for this project

* [ ] Handle when there are no files selected. Use a different message.
* [x] DONE Other UI components? Can I do a table (yes)? I'm going to implement a "fancy list". This is a much
  more "in the weeds" example. I've got something working, I want to refactor it and also add the spinner animation.
   * DONE Functional program
   * DONE (sort of; using a 'domain.go' file) Consider splitting out a "core/domain" package/file
   * DONE (sort of; using a 'domain.go' file) Consider splitting out a "git" package/file
   * Defect. When I press enter on an entry to start fetching its size and then quickly press '/' to start filtering
     and then type some characters, the list filters down to empty even if the filter should match an entry.
   * DONE Go back to value variables (fewer pointers) and allow the Bubble Tea "Update" function to have more logic. The
     experiment was useful, but I want this demo to settle on idiomatic Bubble Tea.
* [ ] Add spinner animation


## Reference

* [GitHub org: `charmbracelet`](https://github.com/charmbracelet)
* [Bubble Tea tutorials: the basics](https://github.com/charmbracelet/bubbletea/tree/master/tutorials/basics)
  * Wow, this was an easy tutorial to follow.
* [GitHub org: `bubbles`](https://github.com/charmbracelet/bubbles)
  * > TUI components for Bubble Tea
* [GitHub org: `lipgloss`](https://github.com/charmbracelet/lipgloss)
  * > Style definitions for nice terminal layout
