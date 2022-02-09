package clang

import (
	"errors"
	"path/filepath"
	"strings"
)

type TidyInvocation struct {
	ExportFile   *string
	DatabaseRoot string
	TargetPath   string
}

// Extract value of CLI option at position int and return updated position.
// If args[position] is one of names, it indicates that the next value is the value of this option. In such case we'll return position+2 and the next value.
// If args[position] starts with one of prefixes, we'll return position+1 and the current value without the prefix.
// In any other case, we'll return position (indicating no shift).
// For example with names = ["-foo", "--foo"], prefixes = ["-f="], any of the following will return value "x":
// ["-foo", "x"], ["--foo", "x"], ["-f=x"]

func ExtractOption(args []string, position int, names []string, prefixes []string) (int, *string) {
	if (position + 1) < len(args) {
		for _, name := range names {
			if args[position] == name {
				value := args[position+1]
				return position + 2, &value
			}
		}
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(args[position], prefix) {
			value := args[position][len(prefix):]
			return position + 1, &value
		}
	}
	return position, nil
}

func ParseTidyCommand(args []string) (*TidyInvocation, error) {
	var invocation TidyInvocation
	for i := 0; i < len(args); {
		if pos, val := ExtractOption(args, i, []string{"-export-fixes", "--export-fixes"}, []string{"--export-fixes="}); pos > i {
			i = pos
			invocation.ExportFile = val
			continue
		}

		if pos, val := ExtractOption(args, i, []string{"-p"}, []string{"-p="}); pos > i {
			i = pos
			invocation.DatabaseRoot = *val
			continue
		}

		if (i + 1) == len(args) {
			invocation.TargetPath = args[i]
		}

		i++
	}

	if len(invocation.TargetPath) == 0 {
		return nil, errors.New("Unable to parse target file path from the clang-tidy command line")
	}
	if len(invocation.DatabaseRoot) == 0 { // if build root is not provided, then clang-tidy defaults to the parent directory of the corresponding file
		invocation.DatabaseRoot = filepath.Dir(invocation.TargetPath)
	}

	return &invocation, nil
}
