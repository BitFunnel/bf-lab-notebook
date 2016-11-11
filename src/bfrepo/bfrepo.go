package bfrepo

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/BitFunnel/LabBook/src/systems/fs"
	"github.com/BitFunnel/LabBook/src/systems/shell"
)

// NOTE: Git remotes are case-insensitive, which is why they're lowercase here.
const bitfunnelHTTPSRemote = `https://github.com/bitfunnel/bitfunnel`
const bitfunnelSSHRemote = `git@github.com:bitfunnel/bitfunnel.git`

// Manager manages the lifecycle of a BitFunnel repository, everything from
// cloning, to checking out a specific version, to building BitFunnel, to
// runinng the REPL.
type Manager interface {
	GetPath() string
	Clone() error
	Fetch() error
	Checkout(revision string) (shell.CmdHandle, error)
	ConfigureBuild() error
	Build() error
	RunFilter(configManifestPath string, samplePath string, sampleArgs []string) error
	RunStatistics(statsManifestPath string, configDir string) error
	RunTermTable(configDir string) error
	RunRepl(configDir string, scriptFile string) error
}

type bfRepoContext struct {
	bitFunnelRoot       string
	buildRoot           string
	bitFunnelExecutable string
}

// New creates a BfRepo object, to manage a BitFunnel repository.
func New(bitFunnelRoot string) Manager {
	buildRoot := filepath.Join(bitFunnelRoot, "build-make")
	bitFunnelExecutable :=
		filepath.Join(buildRoot, "tools", "BitFunnel", "src", "BitFunnel")
	return bfRepoContext{
		bitFunnelRoot:       bitFunnelRoot,
		buildRoot:           buildRoot,
		bitFunnelExecutable: bitFunnelExecutable,
	}
}

// GetPath returns the root path of the BitFunnel repository `repo` manages.
func (repo bfRepoContext) GetPath() string {
	return repo.bitFunnelRoot
}

// Clone clones the canonical GitHub repository, into the folder
// `bitFunnelRoot`.
func (repo bfRepoContext) Clone() (cloneErr error) {
	cloneErr =
		shell.RunCommand("git", "clone", bitfunnelHTTPSRemote, repo.bitFunnelRoot)
	return
}

// Fetch pulls the BitFunnel master from the canonical repository.
func (repo bfRepoContext) Fetch() error {
	chdirHandle, chdirErr := fs.ScopedChdir(repo.bitFunnelRoot)
	if chdirErr != nil {
		return chdirErr
	}
	defer chdirHandle.Dispose()

	originURL, originURLErr :=
		shell.CommandOutput("git", "config", "--get", "remote.origin.url")
	if originURLErr != nil {
		return originURLErr
	}

	lowerOriginURL := strings.ToLower(originURL)

	if lowerOriginURL != bitfunnelSSHRemote &&
		lowerOriginURL != bitfunnelHTTPSRemote {
		return fmt.Errorf("The remote 'origin' in the repository located at "+
			"%s' is required to point at the canonical BitFunnel repository.",
			repo.bitFunnelRoot)
	}

	pullErr := shell.RunCommand("git", "fetch", "origin")
	if pullErr != nil {
		return pullErr
	}
	return nil
}

// Checkout take a path to a canonical BitFunnel repository,
// `bitFunnelRoot`, and checks out a commit from the canonical GitHub
// repository, specified by `sha`.
func (repo bfRepoContext) Checkout(sha string) (shell.CmdHandle, error) {
	chdirHandle, chdirErr := fs.ScopedChdir(repo.bitFunnelRoot)
	if chdirErr != nil {
		return nil, chdirErr
	}
	defer chdirHandle.Dispose()

	// Returns the "short name" of HEAD. Usually this is a branch, like
	// `master`, but if HEAD is detached, it can also simply be `HEAD`.
	headRef, headRefErr :=
		shell.CommandOutput("git", "rev-parse", "--abbrev-ref=strict", "HEAD")
	if headRefErr != nil {
		return nil, headRefErr
	}

	// The commit hash for HEAD.
	headSha, headShaErr := shell.CommandOutput("git", "rev-parse", "HEAD")
	if headShaErr != nil {
		return nil, headShaErr
	}

	// Checkout commit denoted with `sha`.
	checkoutErr := shell.RunCommand("git", "checkout", sha)
	if checkoutErr != nil {
		return nil, checkoutErr
	}

	// Set dispose to reset the head when we're done with it.
	resetHead := func() error {
		chdirHandle, chdirErr := fs.ScopedChdir(repo.bitFunnelRoot)
		if chdirErr != nil {
			return chdirErr
		}
		defer chdirHandle.Dispose()

		var presentRef string
		if headRef == "HEAD" {
			presentRef = headSha
		} else {
			presentRef = headRef
		}

		checkoutErr := shell.RunCommand("git", "checkout", presentRef)
		return checkoutErr
	}

	return shell.MakeHandle(resetHead), nil
}

// Configure switches to the directory of the BitFunnel root, and runs
// the configuration script that generates a makefile.
func (repo bfRepoContext) ConfigureBuild() error {
	chdirHandle, chdirErr := fs.ScopedChdir(repo.bitFunnelRoot)
	if chdirErr != nil {
		return chdirErr
	}
	defer chdirHandle.Dispose()

	configErr := shell.RunCommand("sh", "Configure_Make.sh")
	return configErr
}

// Build switches to the BitFunnel build directory, and builds the code.
func (repo bfRepoContext) Build() error {
	chdirHandle, chdirErr := fs.ScopedChdir(repo.buildRoot)
	if chdirErr != nil {
		return chdirErr
	}
	defer chdirHandle.Dispose()

	buildErr := shell.RunCommand("make", "-j4")
	return buildErr
}

// RunFilter runs the `filter` command in the BitFunnel executable tool.
func (repo bfRepoContext) RunFilter(configManifestPath string, samplePath string, sampleArgs []string) error {
	arguments := []string{
		"filter",
		configManifestPath,
		samplePath,
	}
	arguments = append(
		arguments,
		sampleArgs...)

	return shell.RunCommand(
		repo.bitFunnelExecutable,
		arguments...)
}

// RunStatistics runs the `statistics` command in the BitFunnel executable tool.
func (repo bfRepoContext) RunStatistics(statsManifestPath string, configDir string) error {
	// TODO: Check that this is configured.
	return shell.RunCommand(
		repo.bitFunnelExecutable,
		"statistics",
		statsManifestPath,
		configDir,
		"-text")
}

// RunTermTable runs the `termtable` command in the BitFunnel executable tool.
func (repo bfRepoContext) RunTermTable(configDir string) error {
	return shell.RunCommand(
		repo.bitFunnelExecutable,
		"termtable",
		configDir)
}

// RunRepl runs the BitFunnel repl.
func (repo bfRepoContext) RunRepl(configDir string, scriptFile string) error {
	return shell.RunCommand(
		repo.bitFunnelExecutable,
		"repl",
		configDir,
		"-script",
		scriptFile)
}
