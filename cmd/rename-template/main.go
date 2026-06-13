package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	templateDisplay = "Merc"
	templateSlug    = "ray-d-song"
	templateModule  = "github.com/ray-d-song/merc"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) > 0 && args[0] == "--" {
		args = args[1:]
	}

	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		printUsage(os.Stdout)
		return nil
	}

	if len(args) < 1 || len(args) > 3 {
		printUsage(os.Stderr)
		return errors.New("invalid arguments")
	}

	displayName := args[0]
	slug := ""
	if len(args) >= 2 {
		slug = args[1]
	}
	if slug == "" {
		slug = slugify(displayName)
	}
	if slug == "" {
		return fmt.Errorf("unable to derive slug from display name: %s", displayName)
	}

	goModule := ""
	if len(args) == 3 {
		goModule = args[2]
	}
	if goModule == "" {
		goModule = "github.com/your-org/" + slug
	}

	root, err := repoRoot()
	if err != nil {
		return err
	}

	files, err := textFiles(root)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		fmt.Println("No files to update.")
		return nil
	}

	var goFiles []string
	replacements := []struct {
		old string
		new string
	}{
		{templateModule, goModule},
		{templateDisplay, displayName},
		{templateSlug, slug},
	}

	for _, file := range files {
		changed, err := replaceInFile(file, replacements)
		if err != nil {
			return err
		}
		if changed && strings.HasSuffix(file, ".go") {
			goFiles = append(goFiles, file)
		}
	}

	if len(goFiles) > 0 {
		if err := gofmt(goFiles); err != nil {
			return err
		}
	}

	fmt.Println("Template renamed:")
	fmt.Println("  display:   " + displayName)
	fmt.Println("  slug:      " + slug)
	fmt.Println("  go module: " + goModule)
	return nil
}

func printUsage(out *os.File) {
	fmt.Fprint(out, `Usage:
  go run ./cmd/rename-template -- <display-name> [slug] [go-module]

Examples:
  go run ./cmd/rename-template -- AcmeCI acmeci github.com/acme/acmeci
  go run ./cmd/rename-template -- "Acme Admin"

Defaults:
  slug      lowercases display-name and replaces non-alphanumeric runs with "-"
  go-module github.com/your-org/<slug>
`)
}

func slugify(value string) string {
	value = strings.ToLower(value)
	value = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	return value
}

func repoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(out)), nil
	}

	wd, wdErr := os.Getwd()
	if wdErr != nil {
		return "", wdErr
	}
	return wd, nil
}

func textFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			if shouldSkipDir(root, path) {
				return filepath.SkipDir
			}
			return nil
		}

		if shouldSkipFile(root, path) {
			return nil
		}
		ok, err := isTextFile(path)
		if err != nil {
			return err
		}
		if ok {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func shouldSkipDir(root, path string) bool {
	rel := relative(root, path)
	if rel == "." {
		return false
	}
	parts := strings.Split(filepath.ToSlash(rel), "/")
	switch parts[0] {
	case ".git", "build", "coverage", "dist":
		return true
	}
	for _, part := range parts {
		if part == "node_modules" {
			return true
		}
	}
	return filepath.ToSlash(rel) == "app/infra/static/web-dist"
}

func shouldSkipFile(root, path string) bool {
	rel := filepath.ToSlash(relative(root, path))
	switch rel {
	case "web/pnpm-lock.yaml":
		return true
	}
	return false
}

func relative(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return rel
}

func isTextFile(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	return !bytes.Contains(data, []byte{0}), nil
}

func replaceInFile(path string, replacements []struct {
	old string
	new string
}) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	updated := string(data)
	for _, replacement := range replacements {
		updated = strings.ReplaceAll(updated, replacement.old, replacement.new)
	}
	if updated == string(data) {
		return false, nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return true, os.WriteFile(path, []byte(updated), info.Mode())
}

func gofmt(files []string) error {
	args := append([]string{"-w"}, files...)
	cmd := exec.Command("gofmt", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
