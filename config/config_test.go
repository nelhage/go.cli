package config

import (
	"flag"
	. "launchpad.net/gocheck"
	"strings"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type ConfigSuite struct {
	flags   *flag.FlagSet
	intFlag *int
	strFlag *string
}

var _ = Suite(&ConfigSuite{})

func (s *ConfigSuite) SetUpTest(c *C) {
	s.flags = flag.NewFlagSet("testSuite", flag.ContinueOnError)

	s.intFlag = s.flags.Int("int", 0, "An int-valued flag")
	s.strFlag = s.flags.String("string", "STRING", "A string-valued flag")
}

func (s *ConfigSuite) TestBasic(c *C) {
	err := ParseConfig(s.flags, strings.NewReader(""+
		"# this is a comment\n"+
		"int = 17\n"+
		"\n"+
		"string = hello world\n"))
	c.Assert(err, IsNil)
	c.Assert(*s.intFlag, Equals, 17)
	c.Assert(*s.strFlag, Equals, "hello world")
}

func (s *ConfigSuite) TestNoSuchFlag(c *C) {
	err := ParseConfig(s.flags, strings.NewReader(""+
		"notaflag = 7\n"))
	c.Assert(err.Error(), Equals, "unknown option `notaflag'")
}

func (s *ConfigSuite) TestInvalidLine(c *C) {
	err := ParseConfig(s.flags, strings.NewReader(""+
		"foo \n"))
	c.Assert(err.Error(), Matches, "^illegal config line.*")
}

func (s *ConfigSuite) TestInvalidParse(c *C) {
	err := ParseConfig(s.flags, strings.NewReader(""+
		"int = seventeen\n"))
	c.Assert(err, NotNil)
}

func (s *ConfigSuite) TestWhitespace(c *C) {
	err := ParseConfig(s.flags, strings.NewReader(""+
		"# this is a comment\n"+
		"    int  = 128\n"+
		"\n"+
		"   \t\t \n"+
		"\tstring\t =      value#with spaces\t\t\n"))
	c.Assert(err, IsNil)
	c.Assert(*s.intFlag, Equals, 128)
	c.Assert(*s.strFlag, Equals, "value#with spaces")
}
