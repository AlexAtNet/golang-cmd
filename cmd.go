package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Parse parses the command line into a command and a list of arguments.
func Parse(cmd string) (string, []string) {
	const (
		Inquote       = 1
		InquoteEscape = 2
		InquoteQuote  = 3
		Text          = 4
		TextEscape    = 5
		TextQuote     = 6
		TextSpace     = 7
	)

	items := []string{}
	item := ""
	state := TextSpace
	for _, c := range cmd {
		switch c {
		case ' ':
			switch state {
			case Inquote:
				item += string(c)
			case InquoteEscape:
				item += "\\" + string(c)
				state = Inquote
			case InquoteQuote:
				items = append(items, item)
				item = ""
				state = TextSpace
			case Text:
				items = append(items, item)
				item = ""
				state = TextSpace
			case TextEscape:
				item += string(c)
				state = Text
			case TextQuote:
				item += string(c)
				state = Inquote
			case TextSpace:
				// ignore
			}
		case '"':
			switch state {
			case Inquote:
				state = InquoteQuote
			case InquoteEscape:
				item += string(c)
				state = Inquote
			case InquoteQuote:
				state = Inquote
			case Text:
				state = TextQuote
			case TextEscape:
				item += string(c)
				state = Text
			case TextQuote:
				state = Text
			case TextSpace:
				state = Inquote
			}
		case '\\':
			switch state {
			case Inquote:
				state = InquoteEscape
			case InquoteEscape:
				item += "\\\\"
			case InquoteQuote:
				state = InquoteEscape
			case Text:
				state = TextQuote
			case TextEscape:
				item += "\\"
				state = Text
			case TextQuote:
				state = InquoteEscape
			case TextSpace:
				state = TextEscape
			}
		default:
			switch state {
			case Inquote:
				item += string(c)
			case InquoteEscape:
				item += "\"" + string(c)
			case InquoteQuote:
				state = Text
			case Text:
				item += string(c)
			case TextEscape:
				item += "\\" + string(c)
				state = Text
			case TextQuote:
				item += string(c)
				state = Inquote
			case TextSpace:
				item = string(c)
				state = Text
			}
		}
	}

	switch state {
	case Inquote:
		// autocorrect - add trailing quote
		items = append(items, item)
	case InquoteQuote:
		items = append(items, item)
	case Text:
		items = append(items, item)
	case TextQuote:
		// autocorrect - add trailing quote
		// cmd text" =(autoclose)=> cmd text"" => cmd text
		items = append(items, item)
	case TextSpace:
		// ignore
	}

	if len(items) == 0 {
		return "", []string{}
	}
	return items[0], items[1:]
}

// Run runs cmd and redirects its output to stdout and stderr. If
// the returncode of the cmd is not 0, terminates the current process.
func Run(cmd string) {
	first, args := Parse(cmd)

	p := exec.Command(first, args...)

	stdout, err := p.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
	go func() {
		io.Copy(os.Stdout, stdout)
		stdout.Close()
	}()

	stderr, err := p.StderrPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}

	go func() {
		io.Copy(os.Stderr, stderr)
		stderr.Close()
	}()

	if err := p.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

// Get runs cmd and returns its output. Cmd's stderr is redirected to
// the os.Stderr. If the returncode of the cmd is not 0, terminates
// the current process.
func Get(cmd string) string {
	first, args := Parse(cmd)

	p := exec.Command(first, args...)

	stderr, err := p.StderrPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}

	go func() {
		io.Copy(os.Stderr, stderr)
		stderr.Close()
	}()

	output, err := p.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}

	return string(output)
}
