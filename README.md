# bubbletea-playground

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

Follow these instructions to build and run an example program:

1. Build and run the program:
   * ```shell
     go run .
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


## Wish List

General clean-ups, TODOs and things I wish to implement for this project

* [ ] Handle when there are no files selected. Use a different message.


## Reference

* [GitHub org: `charmbracelet`](https://github.com/charmbracelet)
* [Bubble Tea tutorials: the basics](https://github.com/charmbracelet/bubbletea/tree/master/tutorials/basics)
  * Wow, this was an easy tutorial to follow.
