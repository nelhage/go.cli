package completion

import (
	"flag"
	. "launchpad.net/gocheck"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type CompletionSuite struct{}

var _ = Suite(&CompletionSuite{})

func (s *CompletionSuite) TestParseLine(c *C) {
	testCases := []struct {
		beforePoint string
		afterPoint  string
		words       []string
	}{
		{"hello wo", "rld", []string{"hello", "wo"}},
		{"hello   wo", "rld", []string{"hello", "wo"}},
		{"hello 'there' wo", "rld", []string{"hello", "'there'", "wo"}},
		{"hello 'a b' wo", "rld", []string{"hello", "'a b'", "wo"}},
		{"hello 'a\\ b' wo", "rld", []string{"hello", "'a\\ b'", "wo"}},
		{"hello 'a\\' b' wo", "rld", []string{"hello", "'a\\' b'", "wo"}},
		{`" a string  " with words`, " in it", []string{`" a string  "`, "with", "words"}},
		{"a b ", "", []string{"a", "b", ""}},
		{"pw ", "", []string{"pw", ""}},
	}
	for _, tc := range testCases {
		line := tc.beforePoint + tc.afterPoint
		point := len(tc.beforePoint)
		cl := parseLineForCompletion(line, point)
		c.Check([]string(cl), DeepEquals, tc.words)
	}
}

type FlagCompletionSuite struct {
	flags flag.FlagSet
}

var _ = Suite(&FlagCompletionSuite{})

func (s *FlagCompletionSuite) SetUpSuite(c *C) {
	s.flags.Bool("bool", false, "bool flag")
	s.flags.Int("int", 0, "int flag")
	s.flags.String("str", "", "string flag")
	s.flags.String("str1", "", "string flag 1")
}

func (s *FlagCompletionSuite) TestCompleteFlags(c *C) {
	allFlags := []string{"-bool", "-int", "-str", "-str1"}
	testCases := []struct {
		commandLine []string
		completions []string
		skip        int
	}{
		{[]string{"-"}, allFlags, -1},
		{[]string{""}, allFlags, 0},
		{[]string{"-bool", ""}, allFlags, 1},
		{[]string{"-int", "7", ""}, allFlags, 2},
		{[]string{"-bool", "-str", ""}, []string{}, -1},
		{[]string{"-bool", "-str"}, []string{"-str", "-str1"}, -1},
		{[]string{"-str", "hello", "--int"}, []string{"-int"}, -1},
		{[]string{"-str", "hello", "--int", "42", "", "world"}, nil, 4},
		{[]string{"-str", "hello", "--int", "42", "--", "-str"}, nil, 5},
		{[]string{"-wtf", "-value", ""}, allFlags, 2},
	}
	for _, tc := range testCases {
		var cl CommandLine = append(CommandLine{"cmd"}, tc.commandLine...)
		completions, rest := completeFlags(cl, &s.flags)
		c.Check(completions, DeepEquals, tc.completions)
		if tc.skip < 0 {
			c.Check(rest, IsNil)
		} else {
			c.Check(rest, DeepEquals, cl[tc.skip+1:])
		}
	}
}
