package manifest

import (
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"
	"github.com/myselfBZ/trailpkg/internal/update"
)

var (
	ErrPackageNotFound         = errors.New("Cargo manifest not found on the dock!")
	ErrPackageAlreadyInstalled = errors.New("package already installed")
)

type Package struct {
	DowloadFileName string `json:"download_file_name"`
	Name            string `json:"name"`
	Url             string `json:"url"`
	Outdir          string `json:"out_dir"`

	// OutFiles specifies particular files within Outdir to expose. If set, only
	// these files are used rather than the entire Outdir contents.
	ExportBin      ExportBin       `json:"export_bin"`
	Version        string              `json:"version"`
	BuildSteps     []string            `json:"build_steps"`
	Dependencies   []PackageDependency `json:"dependencies"`
}

type ExportBin struct {
	Bins []string `json:"bins"`
	Dir  string   `json:"dir"`
}

type PackageDependency struct {
	Name            string `json:"name"`
	VersionOrHigher string `json:"version_or_higher"`
	Host            bool   `json:"host"`
}

func NewManifestManager(rootPath string) *ManifestManager {


	manifest := &ManifestManager{
		rootDir:     rootPath,
		manifestDir: path.Join(rootPath, "manifest"),
		binDir:      path.Join(rootPath, "bin"),
		etcDir:      path.Join(rootPath, "etc"),
		storeDir:    path.Join(rootPath, "store"),
		buildDir:    path.Join(rootPath, "build"),
	}

	templateVars := make(map[string]string)
	templateVars["{{TRAIL_STORE}}"] = manifest.storeDir
	templateVars["{{TRAIL_BIN}}"] = manifest.binDir
	templateVars["{{TRAIL_ETC}}"] = manifest.etcDir
	cores := runtime.NumCPU()
	templateVars["{{N_CPUS}}"] = strconv.Itoa(cores)
	manifest.templateVars = templateVars

	manifest.stateManager = NewStateManager(path.Join(rootPath, "state.json"))
	return manifest
}

type ManifestManager struct {
	rootDir              string
	manifestDir          string
	binDir               string
	etcDir               string
	storeDir             string
	buildDir             string
	stateManager *StateManager

	templateVars map[string]string
}

func (m *ManifestManager) UpdateManifest() error {
	err := update.ManifestUpdate(m.manifestDir)
	if err != nil {
		return err
	}

	m.stateManager.UpdateManifestTime(time.Now())
	m.stateManager.Save()
	return nil
}

func (m *ManifestManager) Remove(pckg string) error {
	pkg, err := m.getPackageFromJsonFile(pckg)

	if err != nil {
		return err
	}

	defer func() {
		if err == nil {
			m.stateManager.Remove(pckg)
			m.stateManager.Save()
		} 
	}()

	pkgNameVersion := pkg.Name + "-" + pkg.Version

	storeLocation := path.Join(path.Join(m.storeDir, pkgNameVersion))


	_, found := m.stateManager.Find(pckg)

	if !found {
		return fmt.Errorf("error: package not installed")
	}

	outfiles := pkg.ExportBin.Bins

	if err := os.RemoveAll(storeLocation); err != nil {
		return fmt.Errorf("error deleting store dir of the package: %v", err)
	}

	for _, f := range outfiles {
		if err := os.Remove(path.Join(m.binDir, f)); err != nil {
			return fmt.Errorf("error deleting executables: %v", err)
		}
	}

	return nil
}

func (m *ManifestManager) Install(pckg string) (err error) {

	pkg, err := m.getPackageFromJsonFile(pckg)

	if err != nil {
		return err
	}

	if exists := m.checkIfPkgInstalled(pkg); exists {
		return ErrPackageAlreadyInstalled
	}

	defer func() {
		if err != nil {
			m.cleanupOnFailedInstall(pkg)
		} else {
			m.stateManager.Append(InstallRecord{
				PackageName: pkg.Name,
				Version: pkg.Version,
				InstalledAt: time.Now(),
			})
			m.stateManager.Save()
		}
	}()

	pkgNameVersion := pkg.Name + "-" + pkg.Version


	fmt.Println("Checking for dependencies...")

	missingDeps := m.checkForDependencies(pkg.Dependencies)

	if len(missingDeps) > 0 {
		fmt.Printf("Certain dependencies are not present in the system: %s\n", missingDeps)
		fmt.Println("Installing them...")
	}

	for _, d := range missingDeps {
		fmt.Println("installing dep:", d)
		m.Install(d)
	}

	m.setCurrentPckgNameAndVersion(pkgNameVersion)

	depsWithVersion := m.getPkgVersionMap(pkg.Dependencies)

	// create a build dir for this package
	buildDest := path.Join(m.buildDir, pkgNameVersion)

	if err := os.Mkdir(buildDest, 0755); err != nil {
		return err
	}

	os.Mkdir(path.Join(m.storeDir, pkgNameVersion), 0755)

	fmt.Println("Dowloading...")
	if err := m.downloadSource(pkg.Url, buildDest); err != nil {
		return err
	}

	fileName := pkg.DowloadFileName

	if fileName == "" {
		paths := strings.Split(pkg.Url, "/")
		fileName = paths[len(paths)-1]
	}

	if err := m.extractSource(buildDest, fileName); err != nil {
		return err
	}

	if err := os.Remove(path.Join(buildDest, fileName)); err != nil {
		return err
	}

	if err := m.runBuildCommands(buildDest, pkg.BuildSteps, depsWithVersion); err != nil {
		return err
	}

	binOutDir := pkg.ExportBin.Dir

	if binOutDir == "" {
		binOutDir = pkg.Outdir
	}

	if len(pkg.ExportBin.Bins) > 0 {
		for _, f := range pkg.ExportBin.Bins {
			if err := os.Symlink(
				path.Join(m.storeDir, pkgNameVersion, binOutDir, f),
				path.Join(m.binDir, f),
			); err != nil {
				return fmt.Errorf("error creating a symlink for %s: %v", f, err)
			}
		}
	}

	if err := os.RemoveAll(path.Join(m.buildDir, pkgNameVersion)); err != nil {
		return fmt.Errorf("error cleaning up the build directory: %v", err)
	}

	return nil
}

func (m *ManifestManager) cleanupOnFailedInstall(pkg Package) error {
	pkgNameVersion := pkg.Name + "-" + pkg.Version
	buildDir := path.Join(m.buildDir, pkgNameVersion)
	storeDir := path.Join(m.storeDir, pkgNameVersion)

	if err := os.RemoveAll(buildDir); err != nil {
		return fmt.Errorf("error cleaning up the build dir on failed install: %v", err)
	}

	if err := os.RemoveAll(storeDir); err != nil {
		return fmt.Errorf("error cleaning up the store dir on failed install: %v", err)
	}

	for _, f := range pkg.ExportBin.Bins {
		if err := os.Remove(path.Join(m.binDir, f)); err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("error deleting executables: %v", err)
			}
		}
	}

	return nil
}

// TODO: check the version toooooooo!!!!!
func (m *ManifestManager) checkForDependencies(deps []PackageDependency) []string {
	missing := []string{}

	for _, d := range deps {
		version, err := ParseVersion(d.VersionOrHigher)

		if err != nil {
			fmt.Println("Version coudn't be parsed")
			fmt.Println("FATAL MISTAKEE")
			panic(err)
		}

		if d.Host {
			if ok := m.isDepAvailableHost(d.Name); !ok {
				missing = append(missing, d.Name)
			}
		} else {
			if ok := m.isDepAvailable(d.Name, version); !ok {
				missing = append(missing, d.Name)
			}
		}
	}

	return missing
}

func (m *ManifestManager) setCurrentPckgNameAndVersion(nameVersion string) {
	m.templateVars["{{NAME_AND_VERSION}}"] = nameVersion
}
