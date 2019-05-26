// Package command contains methods called by the CLI to manage
// a mona project.
package command

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/davidsbond/mona/internal/files"
	"github.com/davidsbond/mona/internal/hash"
)

type (
	changeType int
)

const (
	changeTypeBuild changeType = 0
	changeTypeTest  changeType = 1
	changeTypeLint  changeType = 2
)

func getChangedModules(change changeType) ([]*files.ModuleFile, error) {
	lock, err := files.LoadLockFile()

	if err != nil {
		return nil, err
	}

	var out []*files.ModuleFile
	for _, lockInfo := range lock.Modules {
		module, err := files.LoadModuleFile(lockInfo.Location)

		if err != nil {
			return nil, err
		}

		newHash, err := hash.Generate(lockInfo.Location, module.Exclude...)

		if err != nil {
			return nil, err
		}

		diff := false
		switch change {
		case changeTypeBuild:
			diff = lockInfo.BuildHash != newHash
		case changeTypeTest:
			diff = lockInfo.TestHash != newHash
		case changeTypeLint:
			diff = lockInfo.LintHash != newHash
		}

		if diff {
			out = append(out, module)
		}
	}

	return out, nil
}

func streamOutputs(outputs ...io.ReadCloser) {
	for _, output := range outputs {
		go func(o io.ReadCloser) {
			defer o.Close()

			scanner := bufio.NewScanner(o)
			scanner.Split(bufio.ScanLines)

			for scanner.Scan() {
				m := scanner.Text()
				fmt.Println(m)
			}
		}(output)
	}
}

func buildModule(module *files.ModuleFile) error {
	if module.Commands.Build == "" {
		return nil
	}

	parts := strings.Split(module.Commands.Build, " ")
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = module.Location

	return streamCommand(cmd)
}

func lintModule(module *files.ModuleFile) error {
	if module.Commands.Lint == "" {
		return nil
	}

	parts := strings.Split(module.Commands.Lint, " ")
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = module.Location

	return streamCommand(cmd)
}

func testModule(module *files.ModuleFile) error {
	if module.Commands.Test == "" {
		return nil
	}

	parts := strings.Split(module.Commands.Test, " ")
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = module.Location

	return streamCommand(cmd)
}

func streamCommand(cmd *exec.Cmd) error {
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()

	if err != nil {
		return err
	}

	streamOutputs(stdout, stderr)

	if err := cmd.Start(); err != nil {
		return err
	}

	return cmd.Wait()
}

func rangeChangedModules(change changeType, updateHashes bool, fn func(*files.ModuleFile) error) error {
	changed, err := getChangedModules(change)

	if err != nil {
		return err
	}

	newHashes := make(map[string]string)
	for _, module := range changed {
		if err := fn(module); err != nil {
			return err
		}

		if !updateHashes {
			continue
		}

		newHash, err := hash.Generate(module.Location, module.Exclude...)

		if err != nil {
			return err
		}

		newHashes[module.Name+module.Location] = newHash
	}

	if !updateHashes {
		return nil
	}

	lock, err := files.LoadLockFile()

	if err != nil {
		return err
	}

	for i, lockInfo := range lock.Modules {
		if hash, ok := newHashes[lockInfo.Name+lockInfo.Location]; ok {
			switch change {
			case changeTypeBuild:
				lock.Modules[i].BuildHash = hash
			case changeTypeTest:
				lock.Modules[i].TestHash = hash
			}
		}
	}

	return files.UpdateLockFile(lock)
}
