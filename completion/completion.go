package completion

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

var completionLog = log.New(os.Stderr, "completion: ", log.LstdFlags)

type CommandLine []string
type Completer interface {
	Complete(CommandLine) []string
}
type FunctionCompleter func(CommandLine) []string

func (f FunctionCompleter) Complete(cl CommandLine) []string {
	return f(cl)
}

func (c CommandLine) CurrentWord() string {
	return c[len(c)-1]
}

func CompleteIfRequested(completer Completer) {
	if len(os.Args) <= 1 || os.Args[1] != "-do-completion" {
		return
	}
	line := os.Getenv("COMP_LINE")
	pointStr := os.Getenv("COMP_POINT")
	if line == "" || pointStr == "" {
		completionLog.Println("Completion requested, but COMP_LINE and/or COMP_POINT unset.")
		os.Exit(1)
	}

	point, err := strconv.ParseInt(pointStr, 10, 32)
	if err != nil {
		completionLog.Println("Invalid COMP_POINT: ", point)
		os.Exit(1)
	}

	cl := parseLineForCompletion(line, int(point))[1:]

	for _, word := range completer.Complete(cl) {
		fmt.Println(word)
	}
	os.Exit(0)
}

func parseLineForCompletion(line string, point int) CommandLine {
	var cl CommandLine
	var quote rune = 0
	var backslash bool = false
	var word []rune
	for _, char := range line[:point] {
		if backslash {
			word = append(word, char)
			backslash = false
			continue
		}
		if char == '\\' {
			word = append(word, char)
			backslash = true
			continue
		}

		switch quote {
		case 0:
			switch char {
			case '\'', '"':
				word = append(word, char)
				quote = char
			case ' ', '\t':
				if word != nil {
					cl = append(cl, string(word))
				}
				word = nil
			default:
				word = append(word, char)
			}
		case '\'':
			word = append(word, char)
			if char == '\'' {
				quote = 0
			}
		case '"':
			word = append(word, char)
			if char == '"' {
				quote = 0
			}
		}
	}

	return append(cl, string(word))
}

type boolFlag interface {
	flag.Value
	IsBoolFlag() bool
}

func completeFlags(cl CommandLine, flags *flag.FlagSet) (completions []string, rest CommandLine) {
	if len(cl) == 0 {
		return nil, cl
	}
	var inFlag string
	for len(cl) > 1 {
		w := cl[0]
		if inFlag != "" {
			inFlag = ""
		} else if len(w) > 1 && w[0] == '-' && w != "--" {
			if !strings.Contains(w, "=") {
				var i int
				for i = 0; i < len(w) && w[i] == '-'; i++ {
				}
				inFlag = w[i:]
			}
			if flag := flags.Lookup(inFlag); flag != nil {
				if bf, ok := flag.Value.(boolFlag); ok && bf.IsBoolFlag() {
					inFlag = ""
				}
			}
		} else {
			if w == "--" {
				cl = cl[1:]
			}
			return nil, cl
		}
		cl = cl[1:]
	}

	if inFlag != "" {
		// Complete a flag value. No-op for now.
		return []string{}, nil
	} else if len(cl[0]) > 0 && cl[0][0] == '-' {
		// complete a flag name
		prefix := strings.TrimLeft(cl[0], "-")
		flags.VisitAll(func(f *flag.Flag) {
			if strings.HasPrefix(f.Name, prefix) {
				completions = append(completions, "-"+f.Name)
			}
		})
		return completions, nil
	}

	if cl[0] == "" {
		flags.VisitAll(func(f *flag.Flag) {
			completions = append(completions, "-"+f.Name)
		})
	}
	return completions, cl
}

type flagCompleter struct {
	flags *flag.FlagSet
	inner Completer
}

func CompleterWithFlags(flags *flag.FlagSet, completer Completer) Completer {
	return &flagCompleter{
		flags: flags,
		inner: completer,
	}
}

func (c *flagCompleter) Complete(cl CommandLine) []string {
	completions, rest := completeFlags(cl, c.flags)
	if rest != nil {
		if extra := c.inner.Complete(rest); extra != nil {
			completions = append(completions, extra...)
		}
	}

	return completions
}

type setCompleter []string

func (c setCompleter) Complete(cl CommandLine) (completions []string) {
	for _, str := range c {
		if strings.HasPrefix(str, cl.CurrentWord()) {
			completions = append(completions, str)
		}
	}
	return completions
}

func SetCompleter(strs []string) Completer {
	return setCompleter(strs)
}
