package grader

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
)

func RunCommands(commands [][]string, jobDir string, env map[string]string, gid string, stage Stage) {
	fields := make(log.Fields)
	if gid != "" {
		fields["gid"] = gid
	}

	log.WithFields(fields).Infof("Running %s commands", stage)

	for i, argv := range commands {
		fields["command"] = fmt.Sprintf("%s[%d]", stage, i)
		log.WithFields(fields).Info(strings.Join(argv, " "))
		_, _, err := execWithTimeout(argv, jobDir, env, 5*time.Minute)
		if err != nil {
			log.WithFields(fields).Warn(err)
		}
	}
}

func execWithTimeout(argv []string, dir string, env map[string]string, timeout time.Duration) (*bytes.Buffer, int, error) {
	if len(argv) == 0 {
		return nil, 0, errors.New("Empty command")
	}

	expandedArgv := expandArgs(argv, env)
	cmd := exec.Command(expandedArgv[0], expandedArgv[1:]...)

	cmd.Dir = dir
	cmd.Env = buildEnvSlice(env)

	var out bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &out)
	cmd.Stderr = io.MultiWriter(os.Stdout, &out)

	if err := cmd.Start(); err != nil {
		return nil, 0, err
	}
	done := make(chan error)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		if err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					return &out, status.ExitStatus(), nil
				}
			}
			return nil, 0, err
		}
		return &out, 0, nil
	case <-time.After(timeout):
		if err := cmd.Process.Kill(); err != nil {
			return nil, 0, fmt.Errorf("Command timed out (%s), failed to kill process: %v",
				timeout.String(), err)
		}
		return nil, 0, fmt.Errorf("Command timed out (%s), process killed", timeout.String())
	}
}

func expandArgs(argv []string, env map[string]string) []string {
	expandedArgv := make([]string, len(argv))
	for i, arg := range argv {
		expandedArgv[i] = os.Expand(arg, func(key string) string {
			if val, ok := env[key]; ok {
				return val
			}
			return os.Getenv(key)
		})
	}
	return expandedArgv
}

func buildEnvSlice(envMap map[string]string) []string {
	env := os.Environ()
	for key, val := range envMap {
		env = append(env, key+"="+val)
	}
	return env
}
