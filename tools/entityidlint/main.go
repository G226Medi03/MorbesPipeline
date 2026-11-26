// Command entityidlint rejects local MedicalEntity numeric ID references.
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var patterns = []*regexp.Regexp{
	regexp.MustCompile(`\b(?:estatus_id|eaction_id|eactor_id|emedical_procedure_id|medical_procedure_option_id)\b`),
	regexp.MustCompile(`\b(?:status_id|action_id|actor_id|procedure_id|device_id|entity_id|target_entity_id)\b`),
	regexp.MustCompile(`\b(?:StatusID|ActionID|ActorID|ProcedureID|DeviceID|EntityID|SelectedOptionID|MedicalProcedureOptionID)\b`),
	regexp.MustCompile(`"(?:statusId|procedureId|entityId|selectedOptionId)"`),
	regexp.MustCompile(`\b[A-Za-z]*Key\b`),
	regexp.MustCompile(`\b(?:entity_key|target_entity_key|matched_key)\b`),
	regexp.MustCompile(`\bentitykey\b`),
}

func main() {
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "entityidlint: %v\n", err)
		os.Exit(2)
	}
	skipDir := filepath.Clean(filepath.Join(root, "tools", "entityidlint"))
	var violations int
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			name := entry.Name()
			if name == ".git" || name == "bin" || name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
			if filepath.Clean(path) == skipDir {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(entry.Name(), ".go") {
			return nil
		}
		return checkFile(root, path, &violations)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "entityidlint: %v\n", err)
		os.Exit(2)
	}
	if violations > 0 {
		fmt.Printf("entityidlint: %d violation(s)\n", violations)
		os.Exit(1)
	}
	fmt.Println("entityidlint: no violations")
}

func checkFile(root, path string, violations *int) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close() //nolint:errcheck // deferred close

	relative, err := filepath.Rel(root, path)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(file)
	allowFile := false
	for line := 1; line <= 8 && scanner.Scan(); line++ {
		if strings.Contains(scanner.Text(), "entityidlint:allow-file") {
			allowFile = true
		}
	}
	if allowFile {
		_, err := file.Seek(0, 0)
		return err
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		return err
	}
	scanner = bufio.NewScanner(file)
	for line := 1; scanner.Scan(); line++ {
		text := scanner.Text()
		for _, pattern := range patterns {
			if !pattern.MatchString(text) {
				continue
			}
			fmt.Printf("%s:%d: %s\n", filepath.ToSlash(relative), line, strings.TrimSpace(text))
			*violations++
			break
		}
	}
	return scanner.Err()
}
