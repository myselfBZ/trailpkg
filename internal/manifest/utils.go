package manifest

import (
	"encoding/json"
	"regexp"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/myselfBZ/trailpkg/internal/util"
)

func (m *ManifestManager) runBuildCommands(dir string, cmds []string, deps map[string]string) error {
	for _, cmd := range cmds {
		cmd = m.prepareCommand(cmd, deps)
		fmt.Println("RUNNING: ", cmd)
		if err := util.ExecuteCommand(dir, "bash", "-c", cmd); err != nil {
			return fmt.Errorf("build command failed: %v", err)
		}
	}

	return nil
}

func (m *ManifestManager) prepareCommand(cmd string, deps map[string]string) string {
	newCommand := cmd

	for key, val := range m.templateVars {
		newCommand = strings.ReplaceAll(newCommand, key, val)
	}

	var depVersionRegex = regexp.MustCompile(`\{\{DEP_VERSION:(\w+)\}\}`)
	return depVersionRegex.ReplaceAllStringFunc(newCommand, func(match string) string {
        // extract the name from the match
        name := depVersionRegex.FindStringSubmatch(match)[1]
        if version, ok := deps[name]; ok {
            return version
        }
        return match 
    })
}

// we need our own extractor.
func (m *ManifestManager) extractSource(tempDir string, fileName string) error {
	filePath := filepath.Join(tempDir, fileName)

	var cmd *exec.Cmd

	if strings.HasSuffix(fileName, ".tar.gz") || 
	strings.HasSuffix(fileName, ".tgz") || 
	strings.HasSuffix(fileName, ".tar.xz") || 
	strings.HasSuffix(fileName, ".tar.bz2") {
		cmd = exec.Command("tar", "-xf", filePath, "--strip-components=1", "-C", tempDir)
	} else if strings.HasSuffix(fileName, ".zip") {
		cmd = exec.Command("unzip", "-j", filePath, "-d", tempDir)
	}	 else {
		return fmt.Errorf("unsupported archive format: %s", fileName)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (m *ManifestManager) getPackageFromJsonFile(name string) (Package, error) {
	file, err := os.Open(path.Join(m.manifestDir, name+".json"))

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Package{}, ErrPackageNotFound
		} else {
			return Package{}, fmt.Errorf("error opening file: %v", err)
		}
	}

	defer file.Close()

	var pkg Package

	if err := json.NewDecoder(file).Decode(&pkg); err != nil {
		return Package{}, fmt.Errorf("error decoding meta data: %v", err)
	}

	return pkg, nil
}

func (m *ManifestManager) downloadSource(url string, dest string) error {
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}

	if strings.HasSuffix(url, ".git") {
		cmd := exec.Command("git", "clone", url, dest)
		return cmd.Run()
	}

	return m.downloadHTTP(url, dest)
}

func (m *ManifestManager) downloadHTTP(url string, dest string) error {
	fileName := filepath.Base(url)
	filePath := filepath.Join(dest, fileName)

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	return err
}

func (m *ManifestManager) checkIfPkgInstalled(pkg Package) bool {
	// TODO: we should check the stat.json file 
	// But for now we'll just check the dang store dir 
	nameVersion := pkg.Name + "-" + pkg.Version
	location := path.Join(m.storeDir, nameVersion)
	_, err := os.Stat(location)
	return err == nil
}

func (m *ManifestManager) isDepAvailableHost(name string) bool {
    _, err := exec.LookPath(name)
    return err == nil
}


func (m *ManifestManager) isDepAvailable(name string, version Version) bool {
	record, ok := m.stateManager.Find(name)
	if !ok {
		return false
	}

	installedVersion, _ := ParseVersion(record.Version) 
	if !installedVersion.IsAtLeast(version) {
		fmt.Printf("INSTALLED VERSION: %v , REQUIRED VERSION: %v", installedVersion, version)
		return false
	}

	return true
}

func (m *ManifestManager) getPkgVersionMap(deps []PackageDependency) map[string]string {
	result := make(map[string]string)

	for _, d := range deps {
		record, _ := m.stateManager.Find(d.Name)
		result[d.Name] = record.Version
	}

	return result
}

