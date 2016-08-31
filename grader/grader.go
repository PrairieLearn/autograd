package grader

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
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
	jobDir, err := ioutil.TempDir(g.autogradRoot, jobPrefix)
	if err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(jobDir); err != nil {
			log.Printf("Error removing temp dir: %v", err)
		}
	}()

	jobFilePath := filepath.Join(jobDir, jobFileName)
	err = ioutil.WriteFile(jobFilePath, jobData, 0644)
	if err != nil {
		return err
	}
	defer func() {
		if err := os.Remove(jobFilePath); err != nil {
			log.Printf("Error removing temp dir: %v", err)
		}
	}()

	env := map[string]string{
		"AUTOGRAD_GRADER_ROOT": GetGraderRoot(g.autogradRoot),
		"AUTOGRAD_JOB_DIR":     jobDir,
	}

	for i, argv := range g.setupCommands {
		log.Printf("SETUP[%d]: %s", i, strings.Join(argv, " "))
		out, _, err := RunCommand(argv, jobDir, env, 1*time.Minute)
		if err != nil {
			log.Printf("SETUP[%d] error: %v", i, err)
		}
		log.Print(out.String())
	}

	log.Printf("GRADE: %s", strings.Join(g.gradeCommand, " "))
	out, exitCode, err := RunCommand(g.gradeCommand, jobDir, env, g.gradeTimeout)
	if err != nil {
		log.Printf("GRADE error: %v", err)
	}
	log.Print(out.String())
	log.Printf("GRADE score: %d", exitCode)

	for i, argv := range g.cleanupCommands {
		log.Printf("CLEANUP[%d]: %s", i, strings.Join(argv, " "))
		out, _, err := RunCommand(argv, jobDir, env, 1*time.Minute)
		if err != nil {
			log.Printf("CLEANUP[%d] error: %v", i, err)
		}
		log.Print(out.String())
	}

	return nil
}

func GetGraderRoot(autogradRoot string) string {
	return filepath.Join(autogradRoot, graderDir)
}
