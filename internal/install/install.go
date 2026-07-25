// Package install handles installing skills and prompt templates
// into various coding agent harnesses.
package install

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Harness represents a coding agent harness.
type Harness struct {
	Name        string
	Description string
	SkillsDir   func(home string, projectLocal bool) string
	PromptsDir  func(home string, projectLocal bool) string
}

// Supported harnesses.
var Harnesses = map[string]Harness{
	"pi": {
		Name:        "pi",
		Description: "Pi coding agent (pi.dev)",
		SkillsDir: func(home string, projectLocal bool) string {
			if projectLocal {
				return ".pi/skills"
			}
			return filepath.Join(home, ".pi", "agent", "skills")
		},
		PromptsDir: func(home string, projectLocal bool) string {
			if projectLocal {
				return ".pi/prompts"
			}
			return filepath.Join(home, ".pi", "agent", "prompts")
		},
	},
	"claude": {
		Name:        "claude",
		Description: "Claude Desktop / Claude Code",
		SkillsDir: func(home string, projectLocal bool) string {
			if projectLocal {
				return ".claude/skills"
			}
			return filepath.Join(home, ".claude", "skills")
		},
		PromptsDir: nil, // Claude doesn't have prompt templates
	},
	"codex": {
		Name:        "codex",
		Description: "OpenAI Codex CLI",
		SkillsDir: func(home string, projectLocal bool) string {
			if projectLocal {
				return ".codex/skills"
			}
			return filepath.Join(home, ".codex", "skills")
		},
		PromptsDir: nil,
	},
	"cursor": {
		Name:        "cursor",
		Description: "Cursor IDE",
		SkillsDir: func(home string, projectLocal bool) string {
			if projectLocal {
				return ".cursor/skills"
			}
			return filepath.Join(home, ".cursor", "skills")
		},
		PromptsDir: nil,
	},
}

// Install performs the installation.
type Install struct {
	HarnessName  string
	ProjectLocal bool
	SkillsDir    string // source skills directory
	PromptsDir   string // source prompts directory
}

// Result describes what was installed.
type Result struct {
	Harness           string
	Scope             string
	SkillsDir         string
	PromptsDir        string
	Skills            []string
	PromptTemplates   []string
}

// Run performs the installation.
func (inst *Install) Run() (*Result, error) {
	harness, ok := Harnesses[inst.HarnessName]
	if !ok {
		available := make([]string, 0, len(Harnesses))
		for k := range Harnesses {
			available = append(available, k)
		}
		return nil, fmt.Errorf("unknown harness %q (available: %s)", inst.HarnessName, strings.Join(available, ", "))
	}

	home, _ := os.UserHomeDir()
	cwd, _ := os.Getwd()

	skillsDir := harness.SkillsDir(home, inst.ProjectLocal)
	promptsDir := ""
	if harness.PromptsDir != nil {
		promptsDir = harness.PromptsDir(home, inst.ProjectLocal)
	}

	// Make relative paths relative to cwd
	if !filepath.IsAbs(skillsDir) {
		skillsDir = filepath.Join(cwd, skillsDir)
	}
	if promptsDir != "" && !filepath.IsAbs(promptsDir) {
		promptsDir = filepath.Join(cwd, promptsDir)
	}

	result := &Result{
		Harness:    harness.Name,
		Scope:      "global",
		SkillsDir:  skillsDir,
		PromptsDir: promptsDir,
	}
	if inst.ProjectLocal {
		result.Scope = "project-local"
	}

	// Install skills
	if inst.SkillsDir != "" {
		installed, err := copySkillDir(inst.SkillsDir, skillsDir)
		if err != nil {
			return nil, fmt.Errorf("installing skills: %w", err)
		}
		result.Skills = installed
	}

	// Install prompt templates
	if inst.PromptsDir != "" && promptsDir != "" {
		installed, err := copyPromptDir(inst.PromptsDir, promptsDir)
		if err != nil {
			return nil, fmt.Errorf("installing prompt templates: %w", err)
		}
		result.PromptTemplates = installed
	}

	return result, nil
}

// copySkillDir copies skill subdirectories from src to dst.
func copySkillDir(src, dst string) ([]string, error) {
	entries, err := os.ReadDir(src)
	if err != nil {
		return nil, fmt.Errorf("reading source skills dir: %w", err)
	}

	var installed []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillName := entry.Name()
		skillSrc := filepath.Join(src, skillName)
		skillDst := filepath.Join(dst, skillName)

		// Verify it contains SKILL.md
		skillMd := filepath.Join(skillSrc, "SKILL.md")
		if _, err := os.Stat(skillMd); os.IsNotExist(err) {
			continue
		}

		if err := copyDir(skillSrc, skillDst); err != nil {
			return nil, fmt.Errorf("copying skill %s: %w", skillName, err)
		}
		installed = append(installed, skillName)
	}

	return installed, nil
}

// copyPromptDir copies .md prompt templates from src to dst.
func copyPromptDir(src, dst string) ([]string, error) {
	entries, err := os.ReadDir(src)
	if err != nil {
		return nil, fmt.Errorf("reading source prompts dir: %w", err)
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return nil, err
	}

	var installed []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if err := copyFile(srcPath, dstPath); err != nil {
			return nil, fmt.Errorf("copying prompt %s: %w", entry.Name(), err)
		}
		installed = append(installed, entry.Name())
	}

	return installed, nil
}

// copyDir recursively copies a directory.
func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		return copyFile(path, target)
	})
}

// copyFile copies a single file.
func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
