package scan

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"InfraMap/structs"

	"EZDeploy/walker"

	"gopkg.in/yaml.v3"
)

func TestPinnedEZDeployWalkerScansAMonorepoBackend(t *testing.T) {
	root := t.TempDir()
	writeScanFixture(t, root, "apps/api/go.mod", "module api")
	writeScanFixture(t, root, "apps/api/main.go", `package main
import ("os"; "github.com/gin-gonic/gin")
func main() { _ = os.Getenv("DATABASE_URL"); router := gin.Default(); router.GET("/health", health) }`)
	writeScanFixture(t, root, "apps/web/package.json", `{"scripts":{"dev":"next dev"},"dependencies":{"next":"16.2.4"}}`)

	var config walker.Config
	if err := yaml.Unmarshal(ezdeployWalkerConfig, &config); err != nil {
		t.Fatal(err)
	}
	scanner, err := walker.NewScanner(&config)
	if err != nil {
		t.Fatal(err)
	}
	report, err := scanner.Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Services) != 1 || report.Services[0].Root != "apps/api" {
		t.Fatalf("unexpected EZDeploy services: %#v", report.Services)
	}
	analysis := buildEZDeployAnalysis("example/monorepo", "main", "0123456789012345678901234567890123456789", report)
	if len(analysis.Components) != 1 || len(analysis.Dependencies) != 1 || analysis.Dependencies[0].Kind != "database" {
		t.Fatalf("unexpected InfraMap analysis: %#v", analysis)
	}
}

func TestBuildEZDeployAnalysisKeepsScannerEvidence(t *testing.T) {
	report := walker.Report{
		FilesScanned: 12,
		Languages:    map[string]int{"go": 7, "typescript": 5},
		Services: []walker.ServiceCandidate{{
			Name: "go-backend", Runtime: "go", Root: "apps/api", Entry: "apps/api/main.go", Confidence: "high", Evidence: []string{"go.mod", "gin.Default("},
		}},
		RouteHits: []walker.RouteHit{{Method: "GET", Path: "/health", File: "apps/api/main.go"}},
		EnvHits: []walker.EnvHit{
			{Name: "DATABASE_URL", Path: "apps/api/main.go"},
			{Name: "REDIS_URL", Path: "apps/api/main.go"},
		},
	}
	analysis := buildEZDeployAnalysis("example/monorepo", "main", "0123456789012345678901234567890123456789", report)
	if len(analysis.Components) != 1 || analysis.Components[0].Name != "api" || analysis.Components[0].Kind != "backend" {
		t.Fatalf("unexpected components: %#v", analysis.Components)
	}
	if !reflect.DeepEqual(analysis.Components[0].Routes, []string{"/health"}) || !reflect.DeepEqual(analysis.Components[0].Environment, []string{"DATABASE_URL", "REDIS_URL"}) {
		t.Fatalf("scanner evidence was not preserved: %#v", analysis.Components[0])
	}
	if len(analysis.Dependencies) != 2 || analysis.Dependencies[0].Kind != "cache" || analysis.Dependencies[1].Kind != "database" {
		t.Fatalf("unexpected dependencies: %#v", analysis.Dependencies)
	}
}

func TestBuildEZDeployAnalysisSerializesEmptyCollectionsAsArrays(t *testing.T) {
	analysis := buildEZDeployAnalysis("example/empty", "main", "0123456789012345678901234567890123456789", walker.Report{})
	payload, err := json.Marshal(analysis)
	if err != nil {
		t.Fatal(err)
	}
	encoded := string(payload)
	for _, field := range []string{"languages", "components", "dependencies", "dockerfiles"} {
		if strings.Contains(encoded, `"`+field+`":null`) {
			t.Fatalf("%s serialized as null: %s", field, encoded)
		}
	}
}

func TestEZDeployDecisionsMustCoverLatestAnalysis(t *testing.T) {
	analysis := Analysis{
		Components:   []Component{{ID: "component-api"}},
		Dependencies: []Dependency{{ID: "dependency-database"}},
	}
	valid := []structs.EZDeployDecision{
		{ComponentID: "component-api", Action: "deploy", Provider: "aws"},
		{ComponentID: "dependency-database", Action: "connect", Provider: "supabase"},
	}
	if !validEZDeployDecisions(analysis, valid) {
		t.Fatal("expected a complete component plan to be valid")
	}
	if validEZDeployDecisions(analysis, valid[:1]) {
		t.Fatal("expected an incomplete component plan to be rejected")
	}
	duplicate := []structs.EZDeployDecision{valid[0], valid[0]}
	if validEZDeployDecisions(analysis, duplicate) {
		t.Fatal("expected duplicate component decisions to be rejected")
	}
}

func TestExtractRepositoryArchiveRejectsTraversal(t *testing.T) {
	destination := t.TempDir()
	archive := testTarGzip(t, map[string]string{"../outside.txt": "nope"})
	if _, err := extractRepositoryArchive(archive, destination); err == nil {
		t.Fatal("expected path traversal to be rejected")
	}
	if _, err := os.Stat(filepath.Join(destination, "..", "outside.txt")); !os.IsNotExist(err) {
		t.Fatal("unsafe archive wrote outside the extraction directory")
	}
}

func TestExtractRepositoryArchiveReturnsRepositoryRoot(t *testing.T) {
	destination := t.TempDir()
	archive := testTarGzip(t, map[string]string{"repo-sha/apps/api/main.go": "package main"})
	root, err := extractRepositoryArchive(archive, destination)
	if err != nil {
		t.Fatal(err)
	}
	if root != filepath.Join(destination, "repo-sha") {
		t.Fatalf("root = %q, want repository root", root)
	}
}

func testTarGzip(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var compressed bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressed)
	tarWriter := tar.NewWriter(gzipWriter)
	for name, content := range files {
		if err := tarWriter.WriteHeader(&tar.Header{Name: name, Mode: 0o600, Size: int64(len(content)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tarWriter.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	return compressed.Bytes()
}

func writeScanFixture(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
