// Command importlint enforces the A -> B -> C dependency rule.
//
// A feature package (A) must not import an external repository (C) directly
// when an internal adapter package (B) exists for that repository.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type config struct {
	Repos      []repoConfig     `json:"repos"`
	Adapters   []adapterConfig  `json:"adapters"`
	Forbidden  []string         `json:"forbidden"`
	Boundaries []boundaryConfig `json:"boundaries"`
}

type boundaryConfig struct {
	Name string   `json:"name"`
	From string   `json:"from"`
	To   []string `json:"to"`
}

type repoConfig struct {
	Path    string `json:"path"`
	Module  string `json:"module"`
	Enforce bool   `json:"enforce"`
}

type adapterConfig struct {
	Name     string `json:"name"`
	External string `json:"external"`
	Adapter  string `json:"adapter"`
}

type packageInfo struct {
	ImportPath   string
	Name         string
	Dir          string
	Imports      []string
	TestImports  []string
	XTestImports []string
	Module       *struct {
		Path string
	}
}

type violation struct {
	Package   string
	Import    string
	Adapter   adapterConfig
	Forbidden bool
	Rule      string
	Repo      string
}

func main() {
	configPath := flag.String("config", "importlint.json", "path to importlint config")
	reportOnly := flag.Bool("report-only", false, "report violations without failing")
	verbose := flag.Bool("v", false, "print checked repository information")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "importlint: %v\n", err)
		os.Exit(2)
	}

	var violations []violation
	for _, repo := range cfg.Repos {
		packages, err := listPackages(repo.Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "importlint: %s: %v\n", repo.Path, err)
			if repo.Enforce {
				os.Exit(2)
			}
			continue
		}
		if *verbose {
			fmt.Printf("importlint: checked %d packages in %s\n", len(packages), repo.Path)
		}
		for _, pkg := range packages {
			violations = append(violations, checkPackage(repo, pkg, cfg.Adapters, cfg.Forbidden, cfg.Boundaries)...)
		}
	}

	for _, item := range violations {
		if item.Rule != "" {
			fmt.Printf("%s: import %q violates boundary %q\n", item.Package, item.Import, item.Rule)
			continue
		}
		fmt.Printf("%s: import %q is forbidden (A -> C direct reference)\n", item.Package, item.Import)
		if item.Forbidden {
			fmt.Printf("  forbidden module: %s\n", item.Import)
			continue
		}
		fmt.Printf("  use adapter: %s (%s)\n", item.Adapter.Adapter, item.Adapter.External)
	}
	if len(violations) > 0 {
		fmt.Printf("\nimportlint: %d violation(s)\n", len(violations))
		if !*reportOnly {
			os.Exit(1)
		}
		return
	}
	fmt.Println("importlint: no violations")
}

func loadConfig(path string) (config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return config{}, err
	}
	var cfg config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return config{}, err
	}
	if len(cfg.Repos) == 0 {
		return config{}, fmt.Errorf("repos is empty")
	}
	for index := range cfg.Repos {
		if cfg.Repos[index].Path == "" {
			cfg.Repos[index].Path = "."
		}
	}
	return cfg, nil
}

func listPackages(dir string) ([]packageInfo, error) {
	absolute, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command("go", "list", "-json", "./...")
	cmd.Dir = absolute
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(strings.NewReader(string(output)))
	var result []packageInfo
	for decoder.More() {
		var pkg packageInfo
		if err := decoder.Decode(&pkg); err != nil {
			return nil, err
		}
		result = append(result, pkg)
	}
	return result, nil
}

func checkPackage(repo repoConfig, pkg packageInfo, adapters []adapterConfig, forbidden []string, boundaries []boundaryConfig) []violation {
	if !repo.Enforce {
		return nil
	}
	imports := append([]string{}, pkg.Imports...)
	imports = append(imports, pkg.TestImports...)
	imports = append(imports, pkg.XTestImports...)

	var result []violation
	adapterPackage := isAdapterPackage(pkg.ImportPath, adapters)
	for _, importPath := range imports {
		for _, prefix := range forbidden {
			if isExternalImport(importPath, prefix) {
				result = append(result, violation{
					Package:   pkg.ImportPath,
					Import:    importPath,
					Forbidden: true,
					Repo:      repo.Path,
				})
			}
		}
		for _, boundary := range boundaries {
			if !isExternalImport(pkg.ImportPath, boundary.From) {
				continue
			}
			for _, target := range boundary.To {
				if isExternalImport(importPath, target) {
					result = append(result, violation{
						Package: pkg.ImportPath,
						Import:  importPath,
						Rule:    boundary.Name,
						Repo:    repo.Path,
					})
					break
				}
			}
		}
		if adapterPackage {
			continue
		}
		for _, adapter := range adapters {
			if !isExternalImport(importPath, adapter.External) {
				continue
			}
			result = append(result, violation{
				Package: pkg.ImportPath,
				Import:  importPath,
				Adapter: adapter,
				Repo:    repo.Path,
			})
		}
	}
	return result
}

func isAdapterPackage(importPath string, adapters []adapterConfig) bool {
	for _, adapter := range adapters {
		if importPath == adapter.Adapter || strings.HasPrefix(importPath, adapter.Adapter+"/") {
			return true
		}
	}
	return false
}

func isExternalImport(importPath, external string) bool {
	return importPath == external || strings.HasPrefix(importPath, external+"/")
}
