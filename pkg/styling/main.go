// Ad-hoc Lip Gloss styling code.
package main

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

var txt = "Hello, World!"
var multiLineTxt = "Hello,\nWorld!"

var noStyle = lipgloss.NewStyle()

var regularStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("#1693db")).
	Foreground(lipgloss.Color("#ffffff"))

var highlightStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("#e322dd")).
	Foreground(lipgloss.Color("#ffffff")).
	Underline(true)

var paddingStyle = lipgloss.NewStyle().PaddingLeft(2)

var marginStyle = lipgloss.NewStyle().MarginLeft(2)

var leftBorderStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, false, true)

func main() {
	fmt.Printf("%s\n%s\n\n", "Plain text:", txt)

	regularStyledTxt := regularStyle.Render(txt)
	fmt.Printf("%s\n%s\n\n", "Regular style:", regularStyledTxt)

	highlightStyledTxt := highlightStyle.Render(txt)
	fmt.Printf("%s\n%s\n\n", "Highlight style:", highlightStyledTxt)

	runeOnlyHighlightStyledTxt := lipgloss.StyleRunes(txt, []int{1, 3, 5}, noStyle, highlightStyle)
	fmt.Printf("%s\n%s\n\n", "Rune-only highlight style:", runeOnlyHighlightStyledTxt)

	regularPlusRuneHighlightStyledTxt := lipgloss.StyleRunes(txt, []int{1, 3, 5}, regularStyle, highlightStyle)
	fmt.Printf("%s\n%s\n\n", "Regular plus rune highlight style:", regularPlusRuneHighlightStyledTxt)

	paddedTxt := paddingStyle.Render(txt)
	fmt.Printf("%s\n%s\n\n", "Padded text:", paddedTxt)

	paddedPlusRegularStyledTxt := paddingStyle.Inherit(regularStyle).Render(txt)
	fmt.Printf("%s\n%s\n\n", "Padded plus regular style:", paddedPlusRegularStyledTxt)

	// Interestingly, the left margin is also blue. I need to figure out how to combine styles and "renders" as I wish.
	marginPlusRegularStyledTxt := marginStyle.Inherit(regularStyle).Render(txt)
	fmt.Printf("%s\n%s\n\n", "Margin plus regular style:", marginPlusRegularStyledTxt)

	// Regular-styled text and then placed in a margin style. Update: yep worked.
	fmt.Printf("Experiment:\n%s\n\n", marginStyle.Render(regularStyle.Render(txt)))

	// Let's try multi-line. Regular styled.
	fmt.Printf("%s\n%s\n\n", "Multi-line, regular style:", regularStyle.Render(multiLineTxt))

	fmt.Printf("%s\n%s\n\n", "Multi-line, regular inherits margin, then render:", marginStyle.Inherit(regularStyle).Render(multiLineTxt))

	// Render regular style, then place in a margin style.
	fmt.Printf("%s\n%s\n\n", "Multi-line, regular style, then margin style:", marginStyle.Render(regularStyle.Render(multiLineTxt)))

	// Multi-line, regular style, plus rune highlighting
	//
	// There is an unexpected problem. There is space to the right of the first line that's styled with a blue
	// background. Is this the semantic behavior, a defect, or I'm just using it wrong?
	multiLineRegularStylePlusRuneHighlighting := lipgloss.StyleRunes(multiLineTxt, []int{1, 3, 5}, highlightStyle, regularStyle)
	fmt.Printf("%s\n%s\n\n", "Multi-line, rune highlighting:", multiLineRegularStylePlusRuneHighlighting)

	// Put it altogether. Custom customized rune-styling function, plus padding/border.
	multiLineRegularStylePlusMyRuneHighlighting := myStyleRunes(multiLineTxt, []int{1, 3, 5}, highlightStyle, regularStyle)
	fmt.Printf("%s\n%s\n\n", "Multi-line, (my2) rune highlighting, then padding style and left-border style:", paddingStyle.Inherit(leftBorderStyle).Render(multiLineRegularStylePlusMyRuneHighlighting))
}

// Similar to the lipgloss.StyleRunes function, but I found that it doesn't work for multi-line text. I adapted it
// here.
func myStyleRunes(str string, indices []int, matched, unmatched lipgloss.Style) string {
	// Convert slice of indices to a map for easier lookups
	m := make(map[int]struct{})
	for _, i := range indices {
		m[i] = struct{}{}
	}

	noStyle := lipgloss.NewStyle()

	var (
		out   strings.Builder
		group strings.Builder
		style lipgloss.Style
		runes = []rune(str)
	)

	for i, r := range runes {
		if r == '\n' {
			out.WriteString(noStyle.Render("\n"))
			continue
		}

		group.WriteRune(r)

		_, matches := m[i]
		_, nextMatches := m[i+1]

		if matches != nextMatches || i == len(runes)-1 || runes[i+1] == '\n' {
			if matches {
				style = matched
			} else {
				style = unmatched
			}
			out.WriteString(style.Render(group.String()))
			group.Reset()
		}
	}

	return out.String()
}
