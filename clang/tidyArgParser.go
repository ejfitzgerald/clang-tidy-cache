package clang

import (
	"errors"
	"strings"
)

type TidyInvocation struct {
	ExportFile   string
	DatabaseRoot string
	TargetPath   string
}

func ParseTidyCommand(args []string) (*TidyInvocation, error) {
	var invocation TidyInvocation
	for i := 0; i < len(args); {
		if args[i] == "-export-fixes" && (i + 1) < len(args) {
			invocation.ExportFile = args[i + 1]
			i += 2
			continue
		}

		if strings.HasPrefix(args[i], "-p=") {
			invocation.DatabaseRoot = args[i][3:]
		}

		if (i + 1) == len(args) {
			invocation.TargetPath = args[i]
		}

		i++
	}

	if len(invocation.ExportFile) == 0 || len(invocation.DatabaseRoot) == 0 || len(invocation.TargetPath) == 0 {
		return nil, errors.New("Unable to parse incoming tidy command line")
	}

	return &invocation, nil
}
