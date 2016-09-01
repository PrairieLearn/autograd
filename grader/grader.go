package grader

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

const (
	graderDir   = "_grader"
	jobPrefix   = "job_"
	jobFileName = "job_data.json"
)

type Grader struct {
	autogradRoot    string
	setupCommands   [][]string
	gradeCommand    []string
	cleanupCommands [][]string
	gradeTimeout    time.Duration
}

func New(autogradRoot string, setupCommands [][]string, gradeCommand []string, cleanupCommands [][]string,
	gradeTimeout int) *Grader {
	return &Grader{
		autogradRoot:    autogradRoot,
		setupCommands:   setupCommands,
		gradeCommand:    gradeCommand,
		cleanupCommands: cleanupCommands,
		gradeTimeout:    time.Duration(gradeTimeout) * time.Second,
	}
}

func (g *Grader) Grade(jobData []byte) error {
	gid, err := parseGID(jobData)
	if err != nil {
		return errors.New("Error parsing gid from job data")
	}

	jobDir, err := ioutil.TempDir(g.autogradRoot, jobPrefix)
	if err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(jobDir); err != nil {
			log.Warnf("Error removing temp dir: %v", err)
		}
	}()

	jobFilePath := filepath.Join(jobDir, jobFileName)
	err = ioutil.WriteFile(jobFilePath, jobData, 0644)
	if err != nil {
		return err
	}
	defer func() {
		if err := os.Remove(jobFilePath); err != nil {
			log.Warnf("Error removing temp dir: %v", err)
		}
	}()

	env := map[string]string{
		"AUTOGRAD_GRADER_ROOT": GetGraderRoot(g.autogradRoot),
		"AUTOGRAD_JOB_DIR":     jobDir,
	}

	runCommands(g.setupCommands, jobDir, env, gid, "setup")
	runGradeCommand(g.gradeCommand, jobDir, env, gid, g.gradeTimeout)
	runCommands(g.cleanupCommands, jobDir, env, gid, "cleanup")

	return nil
}

func GetGraderRoot(autogradRoot string) string {
	return filepath.Join(autogradRoot, graderDir)
}

func parseGID(jobData []byte) (string, error) {
	var job struct {
		GID string `json:"gid"`
	}
	err := json.Unmarshal(jobData, &job)
	if err != nil {
		return "", err
	}
	return job.GID, nil
}

func runCommands(commands [][]string, jobDir string, env map[string]string, gid, stage string) {
	for i, argv := range commands {
		log.WithFields(log.Fields{
			"gid":  gid,
			"step": fmt.Sprintf("%s[%d]", stage, i),
			"type": "command",
		}).Info(strings.Join(argv, " "))
		out, _, err := RunCommand(argv, jobDir, env, 1*time.Minute)
		if err != nil {
			log.WithFields(log.Fields{
				"gid":  gid,
				"step": fmt.Sprintf("%s[%d]", stage, i),
			}).Warn(err)
		}
		log.WithFields(log.Fields{
			"gid":  gid,
			"step": fmt.Sprintf("%s[%d]", stage, i),
			"type": "output",
		}).Debug(strings.TrimSuffix(out.String(), "\n"))
	}
}

func runGradeCommand(argv []string, jobDir string, env map[string]string, gid string, timeout time.Duration) (
	int, string) {
	log.WithFields(log.Fields{
		"gid":  gid,
		"step": "grade",
		"type": "command",
	}).Info(strings.Join(argv, " "))
	out, exitCode, err := RunCommand(argv, jobDir, env, timeout)
	if err != nil {
		log.WithFields(log.Fields{
			"gid":   gid,
			"stage": "grade",
		}).Warn(err)
	}
	log.WithFields(log.Fields{
		"gid":  gid,
		"step": "grade",
		"type": "output",
	}).Debug(strings.TrimSuffix(out.String(), "\n"))

	log.WithFields(log.Fields{
		"gid":   gid,
		"step":  "grade",
		"score": exitCode,
	}).Info("Grade command exited")

	return exitCode, out.String()
}
