package scan

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"database/sql"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	connections "InfraMap/services/connections"
	workspaces "InfraMap/services/workspaces"
	"InfraMap/structs"

	"EZDeploy/walker"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

const (
	maxGitHubArchiveBytes  = 50 << 20
	maxExtractedBytes      = 250 << 20
	maxExtractedFiles      = 20000
	maxAnalysisComponents  = 48
	maxAnalysisEvidence    = 64
	maxAnalysisDockerfiles = 48
)

//go:embed ezdeploy-walk.yml
var ezdeployWalkerConfig []byte

var (
	githubRepositoryNamePattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)
	githubRevisionPattern       = regexp.MustCompile(`^[0-9a-f]{40}$`)
	gitBranchPattern            = regexp.MustCompile(`^[A-Za-z0-9._/-]+$`)
)

type Analysis struct {
	ID           string                  `json:"id"`
	Repository   string                  `json:"repository"`
	Branch       string                  `json:"branch"`
	Revision     string                  `json:"revision"`
	AnalyzedAt   time.Time               `json:"analyzed_at"`
	FilesScanned int                     `json:"files_scanned"`
	Languages    map[string]int          `json:"languages"`
	Components   []Component             `json:"components"`
	Dependencies []Dependency            `json:"dependencies"`
	Dockerfiles  []walker.DockerfileInfo `json:"dockerfiles"`
	Warnings     []string                `json:"warnings"`
}

type Component struct {
	ID                 string   `json:"id"`
	SourceID           string   `json:"source_id,omitempty"`
	Kind               string   `json:"kind"`
	Name               string   `json:"name"`
	Root               string   `json:"root"`
	Runtime            string   `json:"runtime"`
	Entry              string   `json:"entry"`
	StartCommand       string   `json:"start_command,omitempty"`
	Confidence         string   `json:"confidence"`
	Evidence           []string `json:"evidence"`
	Routes             []string `json:"routes"`
	Environment        []string `json:"environment"`
	Dockerfiles        []string `json:"dockerfiles"`
	SuggestedProviders []string `json:"suggested_providers"`
}

type Dependency struct {
	ID                 string   `json:"id"`
	Kind               string   `json:"kind"`
	Name               string   `json:"name"`
	Evidence           []string `json:"evidence"`
	SuggestedProviders []string `json:"suggested_providers"`
}

func GetEZDeployAnalysis(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, accountID, workspaceID := c.MustGet("userID").(string), c.Param("accountID"), c.Param("workspaceID")
		if _, _, err := workspaces.LoadWorkspace(c, db, userID, accountID, workspaceID); err != nil {
			workspaces.WriteWorkspaceAccessError(c, err)
			return
		}
		resources, err := workspaces.LoadResources(c, db, workspaceID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Workspace sources could not be loaded"})
			return
		}
		source := workspaces.FindResource(resources, "github")
		if source == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Attach a GitHub repository before scanning"})
			return
		}
		analysis, ok := AnalysisFromSource(*source)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "This repository has not been scanned yet"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"analysis": analysis, "decisions": DecisionsFromSource(*source)})
	}
}

func RunEZDeployAnalysis(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, accountID, workspaceID := c.MustGet("userID").(string), c.Param("accountID"), c.Param("workspaceID")
		_, role, err := workspaces.LoadWorkspace(c, db, userID, accountID, workspaceID)
		if err != nil {
			workspaces.WriteWorkspaceAccessError(c, err)
			return
		}
		if role == "viewer" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Viewers cannot run repository scans"})
			return
		}
		resources, err := workspaces.LoadResources(c, db, workspaceID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Workspace sources could not be loaded"})
			return
		}
		source := workspaces.FindResource(resources, "github")
		if source == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Attach a GitHub repository before scanning"})
			return
		}
		repository, branch := source.Name, metadataString(source.Metadata, "default_branch")
		if !githubRepositoryNamePattern.MatchString(repository) || !gitBranchPattern.MatchString(branch) {
			c.JSON(http.StatusConflict, gin.H{"error": "The attached GitHub repository does not have a valid name and branch"})
			return
		}
		token, _, _, err := connections.LoadGitHubConnection(c, db, userID, accountID, source.ConnectionID)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "The GitHub connection could not be opened for scanning"})
			return
		}
		analysis, err := analyzeGitHubRepository(c, token, repository, branch)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "EZDeploy could not scan this repository", "detail": err.Error()})
			return
		}
		updated, err := storeEZDeployAnalysis(c, db, *source, analysis)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "The repository scan could not be saved"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"analysis": analysis, "decisions": []structs.EZDeployDecision{}, "source": updated})
	}
}

func SaveEZDeployDecisions(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, accountID, workspaceID := c.MustGet("userID").(string), c.Param("accountID"), c.Param("workspaceID")
		_, role, err := workspaces.LoadWorkspace(c, db, userID, accountID, workspaceID)
		if err != nil {
			workspaces.WriteWorkspaceAccessError(c, err)
			return
		}
		if role == "viewer" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Viewers cannot change the workspace deployment plan"})
			return
		}
		var request structs.SaveEZDeployDecisionsRequest
		if err := c.ShouldBindJSON(&request); err != nil || len(request.Decisions) > 64 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Enter a valid component plan"})
			return
		}
		resources, err := workspaces.LoadResources(c, db, workspaceID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Workspace sources could not be loaded"})
			return
		}
		source := workspaces.FindResource(resources, "github")
		if source == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Attach and scan a GitHub repository first"})
			return
		}
		analysis, ok := AnalysisFromSource(*source)
		if !ok || !validEZDeployDecisions(analysis, request.Decisions) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "The component plan does not match the latest EZDeploy scan"})
			return
		}
		updated, err := storeEZDeployDecisions(c, db, *source, request.Decisions)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "The component plan could not be saved"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"analysis": analysis, "decisions": request.Decisions, "source": updated})
	}
}

func analyzeGitHubRepository(ctx context.Context, token, repository, branch string) (Analysis, error) {
	archiveRoot, revision, cleanup, err := fetchGitHubArchive(ctx, token, repository, branch)
	if err != nil {
		return Analysis{}, err
	}
	defer cleanup()
	var config walker.Config
	if err := yaml.Unmarshal(ezdeployWalkerConfig, &config); err != nil {
		return Analysis{}, fmt.Errorf("read embedded EZDeploy scanner configuration: %w", err)
	}
	scanner, err := walker.NewScanner(&config)
	if err != nil {
		return Analysis{}, fmt.Errorf("create EZDeploy scanner: %w", err)
	}
	report, err := scanner.Scan(archiveRoot)
	if err != nil {
		return Analysis{}, fmt.Errorf("scan repository: %w", err)
	}
	analysis := buildEZDeployAnalysis(repository, branch, revision, report)
	return analysis, nil
}

func buildEZDeployAnalysis(repository, branch, revision string, report walker.Report) Analysis {
	services := report.Services
	if len(services) > maxAnalysisComponents {
		services = services[:maxAnalysisComponents]
	}
	components := make([]Component, 0, len(services)+len(report.Dockerfiles))
	for _, service := range services {
		routes, environment := report.UniqueRoutePathsForService(service), environmentForRoot(report.EnvHits, service.Root)
		component := Component{
			ID: stableScanID("component", service.Root, service.Entry, service.Runtime), Kind: "backend", Name: componentName(service),
			Root: service.Root, Runtime: service.Runtime, Entry: service.Entry, StartCommand: service.StartCommand,
			Confidence: service.Confidence, Evidence: trimStrings(service.Evidence, maxAnalysisEvidence), Routes: trimStrings(routes, maxAnalysisEvidence), Environment: trimStrings(environment, maxAnalysisEvidence),
			Dockerfiles: trimStrings(dockerfilesForRoot(report.Dockerfiles, service.Root), maxAnalysisEvidence), SuggestedProviders: []string{"aws"},
		}
		components = append(components, component)
	}
	if len(report.Services) == 0 {
		for _, dockerfile := range report.Dockerfiles {
			if len(components) == maxAnalysisComponents {
				break
			}
			root := filepath.ToSlash(filepath.Dir(dockerfile.Path))
			if root == "" {
				root = "."
			}
			name := filepath.Base(root)
			if root == "." {
				name = "Container"
			}
			components = append(components, Component{
				ID: stableScanID("component", root, dockerfile.Path, "docker"), Kind: "container", Name: name, Root: root,
				Runtime: "docker", Confidence: "medium", Evidence: []string{"Dockerfile: " + dockerfile.Path}, Routes: []string{}, Environment: []string{}, Dockerfiles: []string{dockerfile.Path}, SuggestedProviders: []string{"aws"},
			})
		}
	}
	dependencies := inferDependencies(report.EnvHits)
	warnings := []string{"Frontend and migration discovery are limited to capabilities reported by the pinned EZDeploy scanner."}
	if len(report.Services) > len(services) || len(report.Services) == 0 && len(report.Dockerfiles) > len(components) {
		warnings = append(warnings, "The repository contains more component candidates than InfraMap can display in one analysis.")
	}
	if len(components) == 0 {
		warnings = append(warnings, "EZDeploy did not identify a deployable backend or container in this revision.")
	}
	analysisID := stableScanID("analysis", repository, branch, revision)
	return Analysis{
		ID: analysisID, Repository: repository, Branch: branch, Revision: revision, AnalyzedAt: time.Now().UTC(), FilesScanned: report.FilesScanned,
		Languages: nonNilLanguages(report.Languages), Components: components, Dependencies: dependencies, Dockerfiles: trimDockerfiles(report.Dockerfiles, maxAnalysisDockerfiles), Warnings: warnings,
	}
}

func fetchGitHubArchive(ctx context.Context, token, repository, branch string) (string, string, func(), error) {
	client := &http.Client{Timeout: 60 * time.Second}
	revision, err := fetchGitHubRevision(ctx, client, token, repository, branch)
	if err != nil {
		return "", "", func() {}, err
	}
	parts := strings.Split(repository, "/")
	archiveURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/tarball/%s", url.PathEscape(parts[0]), url.PathEscape(parts[1]), url.PathEscape(branch))
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, archiveURL, nil)
	if err != nil {
		return "", "", func() {}, err
	}
	setGitHubArchiveHeaders(request, token)
	response, err := client.Do(request)
	if err != nil {
		return "", "", func() {}, fmt.Errorf("download GitHub archive: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode >= 300 {
		return "", "", func() {}, fmt.Errorf("GitHub archive returned status %d", response.StatusCode)
	}
	compressed, err := io.ReadAll(io.LimitReader(response.Body, maxGitHubArchiveBytes+1))
	if err != nil {
		return "", "", func() {}, err
	}
	if len(compressed) > maxGitHubArchiveBytes {
		return "", "", func() {}, errors.New("repository archive exceeds the 50 MiB scan limit")
	}
	temporary, err := os.MkdirTemp("", "inframap-ezdeploy-scan-")
	if err != nil {
		return "", "", func() {}, err
	}
	cleanup := func() { _ = os.RemoveAll(temporary) }
	root, err := extractRepositoryArchive(compressed, temporary)
	if err != nil {
		cleanup()
		return "", "", func() {}, err
	}
	return root, revision, cleanup, nil
}

func fetchGitHubRevision(ctx context.Context, client *http.Client, token, repository, branch string) (string, error) {
	parts := strings.Split(repository, "/")
	endpoint := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/%s", url.PathEscape(parts[0]), url.PathEscape(parts[1]), url.PathEscape(branch))
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	setGitHubArchiveHeaders(request, token)
	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("resolve GitHub revision: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode >= 300 {
		return "", fmt.Errorf("GitHub revision returned status %d", response.StatusCode)
	}
	var payload struct {
		SHA string `json:"sha"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&payload); err != nil {
		return "", err
	}
	if !githubRevisionPattern.MatchString(payload.SHA) {
		return "", errors.New("GitHub returned an invalid revision")
	}
	return payload.SHA, nil
}

func setGitHubArchiveHeaders(request *http.Request, token string) {
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
}

func extractRepositoryArchive(compressed []byte, destination string) (string, error) {
	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return "", fmt.Errorf("open repository archive: %w", err)
	}
	defer reader.Close()
	archive, archiveEntries, files, extracted := tar.NewReader(reader), 0, 0, int64(0)
	for {
		header, err := archive.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", fmt.Errorf("read repository archive: %w", err)
		}
		archiveEntries++
		if archiveEntries > maxExtractedFiles*2 {
			return "", errors.New("repository archive contains too many entries")
		}
		clean := filepath.Clean(filepath.FromSlash(header.Name))
		if clean == "." || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
			return "", errors.New("repository archive contains an unsafe path")
		}
		target := filepath.Join(destination, clean)
		relative, err := filepath.Rel(destination, target)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return "", errors.New("repository archive escapes the scan directory")
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o700); err != nil {
				return "", err
			}
		case tar.TypeReg, tar.TypeRegA:
			files++
			extracted += header.Size
			if files > maxExtractedFiles || extracted > maxExtractedBytes || header.Size < 0 {
				return "", errors.New("repository exceeds the safe extraction limit")
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
				return "", err
			}
			file, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
			if err != nil {
				return "", err
			}
			_, copyErr := io.CopyN(file, archive, header.Size)
			closeErr := file.Close()
			if copyErr != nil || closeErr != nil {
				return "", errors.Join(copyErr, closeErr)
			}
		}
	}
	entries, err := os.ReadDir(destination)
	if err != nil {
		return "", err
	}
	if len(entries) == 1 && entries[0].IsDir() {
		return filepath.Join(destination, entries[0].Name()), nil
	}
	return destination, nil
}

func componentName(service walker.ServiceCandidate) string {
	root := strings.Trim(service.Root, "/")
	if root != "" && root != "." {
		return filepath.Base(root)
	}
	return service.Name
}

func environmentForRoot(hits []walker.EnvHit, root string) []string {
	root = strings.Trim(strings.TrimSpace(filepath.ToSlash(root)), "/")
	seen := map[string]bool{}
	for _, hit := range hits {
		path := filepath.ToSlash(hit.Path)
		if root == "" || root == "." || path == root || strings.HasPrefix(path, root+"/") {
			seen[hit.Name] = true
		}
	}
	return sortedKeys(seen)
}

func dockerfilesForRoot(files []walker.DockerfileInfo, root string) []string {
	root = strings.Trim(strings.TrimSpace(filepath.ToSlash(root)), "/")
	var matches []string
	for _, file := range files {
		path := filepath.ToSlash(file.Path)
		if root == "" || root == "." || path == root || strings.HasPrefix(path, root+"/") {
			matches = append(matches, path)
		}
	}
	sort.Strings(matches)
	return matches
}

func inferDependencies(hits []walker.EnvHit) []Dependency {
	evidence := map[string]map[string]bool{}
	for _, hit := range hits {
		name := strings.ToUpper(hit.Name)
		kind := ""
		switch {
		case strings.Contains(name, "DATABASE") || strings.Contains(name, "POSTGRES") || strings.HasPrefix(name, "SUPABASE_"):
			kind = "database"
		case strings.Contains(name, "REDIS") || strings.Contains(name, "CACHE_URL"):
			kind = "cache"
		case strings.Contains(name, "S3_") || strings.Contains(name, "STORAGE_BUCKET"):
			kind = "object_storage"
		}
		if kind != "" {
			if evidence[kind] == nil {
				evidence[kind] = map[string]bool{}
			}
			evidence[kind][hit.Name] = true
		}
	}
	providers := map[string][]string{"database": {"supabase", "external"}, "cache": {"aws", "external"}, "object_storage": {"aws", "supabase", "external"}}
	labels := map[string]string{"database": "Database", "cache": "Cache", "object_storage": "Object storage"}
	kinds := make([]string, 0, len(evidence))
	for kind := range evidence {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	dependencies := make([]Dependency, 0, len(kinds))
	for _, kind := range kinds {
		dependencies = append(dependencies, Dependency{
			ID: stableScanID("dependency", kind), Kind: kind, Name: labels[kind], Evidence: sortedKeys(evidence[kind]), SuggestedProviders: providers[kind],
		})
	}
	return dependencies
}

func sortedKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		if key != "" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

func trimStrings(values []string, limit int) []string {
	if len(values) == 0 {
		return []string{}
	}
	if len(values) <= limit {
		return values
	}
	return append([]string(nil), values[:limit]...)
}

func trimDockerfiles(values []walker.DockerfileInfo, limit int) []walker.DockerfileInfo {
	if len(values) == 0 {
		return []walker.DockerfileInfo{}
	}
	if len(values) <= limit {
		return values
	}
	return append([]walker.DockerfileInfo(nil), values[:limit]...)
}

func nonNilLanguages(values map[string]int) map[string]int {
	if values == nil {
		return map[string]int{}
	}
	return values
}

func stableScanID(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return parts[0] + "-" + hex.EncodeToString(sum[:8])
}

func metadataString(metadata map[string]interface{}, key string) string {
	value, _ := metadata[key].(string)
	return value
}

func AnalysisFromSource(resource structs.WorkspaceResource) (Analysis, bool) {
	payload, ok := resource.Metadata["ezdeploy_analysis"]
	if !ok {
		return Analysis{}, false
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return Analysis{}, false
	}
	var analysis Analysis
	if err := json.Unmarshal(encoded, &analysis); err != nil || analysis.ID == "" {
		return Analysis{}, false
	}
	return analysis, true
}

func DecisionsFromSource(resource structs.WorkspaceResource) []structs.EZDeployDecision {
	payload, ok := resource.Metadata["ezdeploy_decisions"]
	if !ok {
		return []structs.EZDeployDecision{}
	}
	encoded, _ := json.Marshal(payload)
	var decisions []structs.EZDeployDecision
	if json.Unmarshal(encoded, &decisions) != nil {
		return []structs.EZDeployDecision{}
	}
	return decisions
}

func validEZDeployDecisions(analysis Analysis, decisions []structs.EZDeployDecision) bool {
	available, seen := map[string]bool{}, map[string]bool{}
	for _, component := range analysis.Components {
		available[component.ID] = true
	}
	for _, dependency := range analysis.Dependencies {
		available[dependency.ID] = true
	}
	validActions := map[string]bool{"connect": true, "deploy": true, "existing": true, "later": true}
	validProviders := map[string]bool{"": true, "aws": true, "vercel": true, "supabase": true, "external": true}
	for _, decision := range decisions {
		if !available[decision.ComponentID] || seen[decision.ComponentID] || !validActions[decision.Action] || !validProviders[decision.Provider] {
			return false
		}
		if decision.Action != "later" && decision.Provider == "" {
			return false
		}
		seen[decision.ComponentID] = true
	}
	return len(seen) == len(available)
}

func storeEZDeployAnalysis(ctx context.Context, db *sql.DB, source structs.WorkspaceResource, analysis Analysis) (structs.WorkspaceResource, error) {
	payload, err := json.Marshal(analysis)
	if err != nil {
		return structs.WorkspaceResource{}, err
	}
	_, err = db.ExecContext(ctx, `UPDATE workspace_resources SET metadata = metadata || jsonb_build_object(
		'ezdeploy_analysis', $1::jsonb, 'ezdeploy_decisions', '[]'::jsonb
	), updated_at = NOW() WHERE id = $2`, string(payload), source.ID)
	if err != nil {
		return structs.WorkspaceResource{}, err
	}
	return reloadSource(ctx, db, source.WorkspaceID, source.ID)
}

func storeEZDeployDecisions(ctx context.Context, db *sql.DB, source structs.WorkspaceResource, decisions []structs.EZDeployDecision) (structs.WorkspaceResource, error) {
	payload, err := json.Marshal(decisions)
	if err != nil {
		return structs.WorkspaceResource{}, err
	}
	_, err = db.ExecContext(ctx, `UPDATE workspace_resources SET metadata = metadata || jsonb_build_object(
		'ezdeploy_decisions', $1::jsonb
	), updated_at = NOW() WHERE id = $2`, string(payload), source.ID)
	if err != nil {
		return structs.WorkspaceResource{}, err
	}
	return reloadSource(ctx, db, source.WorkspaceID, source.ID)
}

func reloadSource(ctx context.Context, db *sql.DB, workspaceID, sourceID string) (structs.WorkspaceResource, error) {
	resources, err := workspaces.LoadResources(ctx, db, workspaceID)
	if err != nil {
		return structs.WorkspaceResource{}, err
	}
	for _, resource := range resources {
		if resource.ID == sourceID {
			return resource, nil
		}
	}
	return structs.WorkspaceResource{}, sql.ErrNoRows
}
