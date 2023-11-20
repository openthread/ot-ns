// Copyright (c) 2023, The OTNS Authors.
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
// 3. Neither the name of the copyright holder nor the
//    names of its contributors may be used to endorse or promote products
//    derived from this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package cli

import (
	_ "embed"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/mitchellh/go-wordwrap"
	"github.com/openthread/ot-ns/logger"
	"golang.org/x/term"
)

type Help struct {
	termWidth     uint
	maxCmdWidth   uint
	commands      map[string]string
	commandsShort map[string]string
}

var (
	cmdHeaderPattern  = regexp.MustCompile("^### .+")
	linkTargetPattern = regexp.MustCompile(`\(#[a-z]+\)`)
)

// Embed the CLI help file as a static resource.
//
//go:embed README.md
var cliHelpFile string

// Creates new Help object. It is used to display CLI commands help to the user.
func newHelp() Help {
	h := Help{
		termWidth:     80,
		maxCmdWidth:   10,
		commands:      make(map[string]string),
		commandsShort: make(map[string]string),
	}
	h.parseHelpFile()
	h.update()
	return h
}

// Updates the Help object to take into account current user's terminal size.
func (help *Help) update() {
	fdTerm := int(os.Stdout.Fd()) // Windows platform requires cast to int.
	if term.IsTerminal(fdTerm) {
		width, _, err := term.GetSize(fdTerm)
		logger.PanicIfError(err, "Could not get terminal size.")
		help.termWidth = uint(width)
	}
}

// Output short help for all commands.
func (help *Help) outputGeneralHelp() string {
	cmdHelp := ""
	// get a sorted list of commands
	cmds := make([]string, 0, len(help.commandsShort))
	for k := range help.commandsShort {
		cmds = append(cmds, k)
	}
	sort.Strings(cmds)

	for _, c := range cmds {
		cmdHelp += fmt.Sprintf("%-15s %s\n", c, help.commandsShort[c])
	}
	return cmdHelp +
		wordwrap.WrapString("\nFor detailed help per command, use: 'help <command>'\n",
			help.termWidth) +
		wordwrap.WrapString("\nFor detailed CLI command reference in browser go to:\n"+
			"https://github.com/EskoDijk/ot-ns/blob/main/cli/README.md\n",
			help.termWidth)
}

// Output help for one specific command.
func (help *Help) outputCommandHelp(command string) string {
	return help.outputHelp([]string{command})
}

// Output help for one or more specific commands, in given order.
func (help *Help) outputHelp(commands []string) string {
	help.update()
	s := ""
	for _, cmd := range commands {
		explanation, ok := help.commands[cmd]
		if !ok {
			explanation = "(Non-existent command.)"
		}
		w := help.termWidth - help.maxCmdWidth - 1
		explWrapped := strings.Split(wordwrap.WrapString(explanation, w), "\n")
		for _, line := range explWrapped {
			if cmdHeaderPattern.MatchString(line) {
				s += line[strings.Index(line, " ")+1:] + "\n"
			} else {
				s += "  " + line + "\n"
			}
		}
	}
	return s
}

func (help *Help) parseHelpFile() {
	indentString := "    "
	lines := strings.Split(cliHelpFile, "\n")
	activeCmd := ""
	indent := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if len(line) == 0 {
			continue
		}

		if line == "```bash" {
			line = "\nExample:"
			indent = 2
		} else if line == "```shell" {
			line = "\nDefinition:"
			indent = 2
		} else if line == "```" {
			line = ""
			indent = 0
		} else if cmdHeaderPattern.MatchString(line) {
			activeCmd = strings.TrimSpace(line[strings.Index(line, " ")+1:])
			help.commands[activeCmd] = ""
			help.commandsShort[activeCmd] = ""
			line = activeCmd
			indent = 0
		}

		if len(activeCmd) > 0 {
			help.commands[activeCmd] += indentString[0:indent] + markdownUnquote(line) + "\n"
			if line != activeCmd && len(help.commandsShort[activeCmd]) == 0 {
				firstSentence := line
				idx := strings.Index(line, ".")
				if idx > 0 {
					firstSentence = line[:idx+1]
				}
				help.commandsShort[activeCmd] = firstSentence
			}
		}
	}
}

func markdownUnquote(md string) string {
	// TODO: consider that double backslash may be present in the future in the Markdown.
	// TODO: change MD links to text
	md = strings.ReplaceAll(md, "\\", "")
	md = linkTargetPattern.ReplaceAllString(md, "")
	return md
}
