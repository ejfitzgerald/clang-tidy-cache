package clang

import (
	"crypto/sha256"
	"errors"
	"github.com/google/shlex"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

type CompilerCommand struct {
	Compiler   string
	Arguments  []string
	OutputPath string
	InputPath  string
}

func ParseClangCommandString(commands string) (*CompilerCommand, error) {
	words, err := shlex.Split(commands)
	if err != nil {
		return nil, err
	}

	var cmd CompilerCommand
	cmd.Compiler = words[0]
	cmd.Arguments = make([]string, 0, len(words))

	// strip the compiler from the front
	words = words[1:]

	for i := 0; i < len(words); {
		if words[i] == "-c" && (i+1) < len(words) {
			cmd.InputPath = words[i+1]
			i += 2
			continue
		}

		if words[i] == "-o" && (i+1) < len(words) {
			cmd.OutputPath = words[i+1]
			i += 2
			continue
		}

		// all other arguments are just passed to the argument list
		cmd.Arguments = append(cmd.Arguments, words[i])
		i++
	}

	if len(cmd.InputPath) == 0 || len(cmd.OutputPath) == 0 {
		return nil, errors.New("Unable to determine input or output path")
	}

	return &cmd, nil
}

func EvaluatePreprocessedFile(buildRoot string, command *CompilerCommand) ([]byte, error) {
	// make the temporary file
	tmpfile, err := ioutil.TempFile("", "ctc-")
	if err != nil {
		return nil, err
	}

	// cache the filename
	filename := tmpfile.Name()

	// close down the file
	err = tmpfile.Close()
	if err != nil {
		return nil, err
	}

	// build up all of the args
	args := make([]string, 0, len(command.Arguments)+10)
	args = append(args, command.Arguments...)
	args = append(args, "-E", "-o", filename, command.InputPath)

	// run the preprocessor
	cmd := exec.Command(command.Compiler, args...)
	cmd.Dir = buildRoot
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	// read the contents of the file am hash it
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return nil, err
	}

	// compute the final digest
	digest := hasher.Sum(nil)

	// remove the file (clean up)
	err = os.Remove(filename)
	if err != nil {
		return nil, err
	}

	return digest, nil
}
