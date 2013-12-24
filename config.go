package config

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func LoadConfig(flags *flag.FlagSet, basename string) error {
	path := os.ExpandEnv(fmt.Sprintf("${HOME}/.%s", basename))
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()
	return ParseConfig(flags, f)
}

func ParseConfig(flags *flag.FlagSet, f io.Reader) error {
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		bits := strings.SplitN(line, "=", 2)
		if len(bits) != 2 {
			return fmt.Errorf("Illegal config line: `%s'", line)
		}

		key := strings.TrimSpace(bits[0])
		value := strings.TrimSpace(bits[1])

		if flag := flags.Lookup(key); flag == nil {
			return fmt.Errorf("Unknown option `%s'", bits[0])
		}

		if err := flags.Set(key, value); err != nil {
			return err
		}
	}
	return nil
}
