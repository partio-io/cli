// Package programshape checks that a minions program keeps its instructions
// where the program parser can reach them.
//
// The parser keeps four things from a program body: the H1 heading with the
// prose directly under it, a ## Context section, a ## Planner section, and a
// ## Agents section with its H3 sub-sections. It discards every other heading
// and its body, silently. When a program defines no agents, the executor
// synthesizes one whose entire instruction set is the prose under the H1.
//
// A program that puts its procedure in a discarded section therefore runs on
// one or two sentences and improvises the rest.
package programshape

import (
	"regexp"
	"strings"
)

// keptSections are the H2 headings the parser reads. The comparison is on the
// lower-case heading text, without the leading "## ".
var keptSections = map[string]bool{
	"context": true,
	"planner": true,
	"agents":  true,
}

// Finding is one discarded section whose body tells the agent what to do.
type Finding struct {
	// Heading is the section heading as written, for example "## Steps".
	Heading string
	// Reason names the instruction content found, for the failure message.
	Reason string
}

// Check reports the sections whose instructions never reach the model.
//
// Check returns nil for a program that defines an ## Agents section, because
// the parser reads an agent's prose in full. It also returns nil for a program
// whose entire instruction set is the prose under the H1, because that text
// does reach the model.
func Check(src string) []Finding {
	sections := topLevelSections(stripFrontmatter(src))

	for _, s := range sections {
		if headingText(s.heading) == "agents" {
			return nil
		}
	}

	var findings []Finding
	for _, s := range sections {
		if keptSections[headingText(s.heading)] {
			continue
		}
		if reason := instructionContent(s.body); reason != "" {
			findings = append(findings, Finding{Heading: s.heading, Reason: reason})
		}
	}
	return findings
}

// section is one H2 heading and the body under it.
type section struct {
	heading string
	body    string
}

// stripFrontmatter removes a leading YAML frontmatter block, which the parser
// reads separately from the Markdown body.
func stripFrontmatter(src string) string {
	lines := strings.Split(src, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return src
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return strings.Join(lines[i+1:], "\n")
		}
	}
	return src
}

// topLevelSections splits a program body on its H2 headings. Text before the
// first H2, which is the H1 and the prose under it, belongs to no section and
// is dropped here because it always reaches the model.
func topLevelSections(body string) []section {
	var (
		sections  []section
		current   *section
		bodyLines []string
		inFence   bool
	)

	flush := func() {
		if current == nil {
			return
		}
		current.body = strings.Join(bodyLines, "\n")
		sections = append(sections, *current)
	}

	for line := range strings.SplitSeq(body, "\n") {
		if isFence(line) {
			inFence = !inFence
		} else if !inFence && strings.HasPrefix(line, "## ") {
			flush()
			current = &section{heading: strings.TrimSpace(line)}
			bodyLines = nil
			continue
		}
		if current != nil {
			bodyLines = append(bodyLines, line)
		}
	}
	flush()

	return sections
}

// headingText returns the lower-case heading text, without its "##" marker.
func headingText(heading string) string {
	return strings.ToLower(strings.TrimSpace(strings.TrimPrefix(heading, "##")))
}

// numberedItem matches a numbered list item, which reads as one step of a
// procedure.
var numberedItem = regexp.MustCompile(`(?m)^\s*\d+\.\s+\S`)

// instructionContent names the instruction content in a section body, or
// returns an empty string when the body only explains, attributes, or asserts.
// A numbered step list and a fenced block are the two markers that separate a
// procedure the agent must follow from prose about why the work matters.
func instructionContent(body string) string {
	if hasFence(body) {
		return "a command or code block"
	}
	if numberedItem.MatchString(body) {
		return "a numbered step list"
	}
	return ""
}

// hasFence reports whether the body opens a fenced code block.
func hasFence(body string) bool {
	for line := range strings.SplitSeq(body, "\n") {
		if isFence(line) {
			return true
		}
	}
	return false
}

// isFence reports whether a line opens or closes a fenced code block.
func isFence(line string) bool {
	trimmed := strings.TrimLeft(line, " \t")
	return strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~")
}
