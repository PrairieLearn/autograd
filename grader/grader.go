package grader

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

type Stage string

const (
	graderDir   = "_grader"
	jobPrefix   = "job_"
	jobFileName = "job_data.json"

	InitStage    Stage = "init"
	SetupStage   Stage = "setup"
	GradeStage   Stage = "grade"
	CleanupStage Stage = "cleanup"
)

type Grader struct {
	autogradRoot    string
	setupCommands   [][]string
	gradeCommand    []string
	cleanupCommands [][]string
	gradeTimeout    time.Duration
}

type Result struct {
	GID     string  `json:"gid"`
	Grading Grading `json:"grading"`
}

type Grading struct {
	Score    int    `json:"score"`
	Feedback []byte `json:"feedback"`
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

func (g *Grader) Grade(jobData []byte) (*Result, error) {
	gid, err := parseGID(jobData)
	if err != nil {
		return nil, errors.New("Error parsing gid from job data")
	}

	jobDir, err := ioutil.TempDir(g.autogradRoot, jobPrefix)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := os.RemoveAll(jobDir); err != nil {
			log.Warnf("Error removing temp dir: %v", err)
		}
	}()

	jobFilePath := filepath.Join(jobDir, jobFileName)
	err = ioutil.WriteFile(jobFilePath, jobData, 0644)
	if err != nil {
		return nil, err
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

	RunCommands(g.setupCommands, jobDir, env, gid, "setup")
	score, feedback := runGradeCommand(g.gradeCommand, jobDir, env, gid, g.gradeTimeout)
	RunCommands(g.cleanupCommands, jobDir, env, gid, "cleanup")

	return &Result{
		GID: gid,
		Grading: Grading{
			Score:    score,
			Feedback: feedback,
		},
	}, nil
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

func runGradeCommand(argv []string, jobDir string, env map[string]string, gid string, timeout time.Duration) (
	int, []byte) {
	log.WithFields(log.Fields{
		"gid": gid,
	}).Infof("Running grade command")

	log.WithFields(log.Fields{
		"gid":  gid,
		"step": GradeStage,
	}).Info(strings.Join(argv, " "))
	out, exitCode, err := execWithTimeout(argv, jobDir, env, timeout)
	if err != nil {
		log.WithFields(log.Fields{
			"gid":   gid,
			"stage": "grade",
		}).Warn(err)
	}

	log.WithFields(log.Fields{
		"gid":   gid,
		"step":  GradeStage,
		"score": exitCode,
	}).Info("Grade command exited")

	return exitCode, out.Bytes()
}
