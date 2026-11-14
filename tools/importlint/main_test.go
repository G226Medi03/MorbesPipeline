package main

import "testing"

func TestCheckPackageRejectsForbiddenImport(t *testing.T) {
	pkg := packageInfo{
		ImportPath: "github.com/G226Medi03/MorbesPipeline/internal/pipeline",
		Imports:    []string{"github.com/G226Medi03/MedicalEntity/ename/estatus"},
	}
	violations := checkPackage(
		repoConfig{Path: ".", Enforce: true},
		pkg,
		nil,
		[]string{"github.com/G226Medi03/MedicalEntity"},
	)
	if len(violations) != 1 || !violations[0].Forbidden {
		t.Fatalf("violations = %+v, want one forbidden violation", violations)
	}
}

func TestCheckPackageRejectsForbiddenTestImport(t *testing.T) {
	pkg := packageInfo{
		ImportPath:  "github.com/G226Medi03/MorbesPipeline/internal/pipeline",
		TestImports: []string{"github.com/G226Medi03/MedicalEntity/ename/estatus"},
	}
	violations := checkPackage(
		repoConfig{Path: ".", Enforce: true},
		pkg,
		nil,
		[]string{"github.com/G226Medi03/MedicalEntity"},
	)
	if len(violations) != 1 || !violations[0].Forbidden {
		t.Fatalf("violations = %+v, want one forbidden violation", violations)
	}
}

func TestCheckPackageAllowsAdapterImport(t *testing.T) {
	adapter := adapterConfig{
		Name:     "pipelinecore",
		External: "github.com/G226Medi03/PipeLineCore",
		Adapter:  "github.com/G226Medi03/MorbesPipeline/internal/external/pipelinecore",
	}
	pkg := packageInfo{
		ImportPath: adapter.Adapter,
		Imports:    []string{"github.com/G226Medi03/PipeLineCore"},
	}
	violations := checkPackage(
		repoConfig{Path: ".", Enforce: true},
		pkg,
		[]adapterConfig{adapter},
		nil,
	)
	if len(violations) != 0 {
		t.Fatalf("violations = %+v, want 0", violations)
	}
}

func TestCheckPackageRejectsAdapterBypass(t *testing.T) {
	adapter := adapterConfig{
		Name:     "pipelinecore",
		External: "github.com/G226Medi03/PipeLineCore",
		Adapter:  "github.com/G226Medi03/MorbesPipeline/internal/external/pipelinecore",
	}
	pkg := packageInfo{
		ImportPath: "github.com/G226Medi03/MorbesPipeline/internal/pipeline",
		Imports:    []string{"github.com/G226Medi03/PipeLineCore"},
	}
	violations := checkPackage(
		repoConfig{Path: ".", Enforce: true},
		pkg,
		[]adapterConfig{adapter},
		nil,
	)
	if len(violations) != 1 || violations[0].Forbidden {
		t.Fatalf("violations = %+v, want one adapter violation", violations)
	}
}
