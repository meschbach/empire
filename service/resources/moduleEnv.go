package resources

import (
	"errors"
	"fmt"
	"github.com/go-git/go-billy/v5"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type CapturingPackageManager struct {
	InstallPackages []string
}

func (c *CapturingPackageManager) EnsureInstalled(name string) error {
	c.InstallPackages = append(c.InstallPackages, name)
	return nil
}

type resourceType struct {
	resources map[string]*resource
}

type resource struct {
}

type ModuleEnv struct {
	resourceTypes    map[string]*resourceType
	source           billy.Filesystem
	sourceWorkingDir string
	target           billy.Filesystem
	targetWorkingDir string
	Runtime          *CapturingPackageManager
}

func NewModuleEnv(source billy.Filesystem, sourceWorkingDir string, target billy.Filesystem, targetWorkingDir string, runtime *CapturingPackageManager) *ModuleEnv {
	return &ModuleEnv{
		resourceTypes:    make(map[string]*resourceType),
		source:           source,
		sourceWorkingDir: sourceWorkingDir,
		target:           target,
		targetWorkingDir: targetWorkingDir,
		Runtime:          runtime,
	}
}

func (a *ModuleEnv) HasCompleted(resourceType, resourceName string) bool {
	if names, ok := a.resourceTypes[resourceType]; !ok {
		return false
	} else {
		if _, ok := names.resources[resourceName]; !ok {
			return false
		}
		return true
	}
}

func (a *ModuleEnv) Completed(resourceTypeName, resourceName string) bool {
	if _, ok := a.resourceTypes[resourceTypeName]; !ok {
		a.resourceTypes[resourceTypeName] = &resourceType{make(map[string]*resource)}
	}
	names := a.resourceTypes[resourceTypeName]
	if _, ok := names.resources[resourceName]; !ok {
		names.resources[resourceName] = &resource{}
		return true
	}
	return false
}

func (a *ModuleEnv) ResolveConfigFile(element string) (string, error) {
	return filepath.Join(a.sourceWorkingDir, element), nil
}
func (a *ModuleEnv) ResolveAbsolutePath(element string) (string, error) {
	return filepath.Join(a.targetWorkingDir, element), nil
}
func (a *ModuleEnv) EnsureFileExists(path string) (bool, error) {
	if _, err := a.target.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func (a *ModuleEnv) MakeDirectory(path string) error {
	return a.target.MkdirAll(path, 0744)
}

type CopyFileError struct {
	SourceFilePath      string
	DestinationFilePath string
	Problem             error
}

func (c *CopyFileError) Error() string {
	return fmt.Sprintf("Failed to copy file %q to %q because %s", c.SourceFilePath, c.DestinationFilePath, c.Problem)
}

type CopyOpt struct {
}

func (a *ModuleEnv) CopyFile(source string, target string, copy CopyOpt) error {
	src, err := a.source.Open(source)
	if err != nil {
		return &CopyFileError{
			SourceFilePath:      source,
			DestinationFilePath: target,
			Problem:             err,
		}
	}
	defer src.Close()

	dst, err := a.target.Create(target)
	if err != nil {
		return &CopyFileError{
			SourceFilePath:      source,
			DestinationFilePath: target,
			Problem:             err,
		}
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return &CopyFileError{
			SourceFilePath:      source,
			DestinationFilePath: target,
			Problem:             err,
		}
	}
	return nil
}

func (a *ModuleEnv) resolve(cfg *File) error {
	var problems []error
	for _, dir := range cfg.Directory {
		if err := dir.Perform(a); err != nil {
			fmt.Printf("Dir failed because %s\n", err.Error())
		}
	}

	for _, file := range cfg.StaticLocalFiles {
		if err := file.Perform(a); err != nil {
			problems = append(problems, &ResourceError{
				Type:    "static_local_file",
				Name:    file.Target,
				Problem: err,
			})
		}
	}

	for _, gitModule := range cfg.GitModule {
		if err := gitModule.Perform(a); err != nil {
			problems = append(problems, &ResourceError{
				Type:    "git_module",
				Name:    gitModule.Name,
				Problem: err,
			})
		}
	}

	for _, pkgs := range cfg.Package {
		if err := pkgs.Perform(a); err != nil {
			problems = append(problems, &ResourceError{
				Type:    "package",
				Name:    pkgs.Name,
				Problem: err,
			})
		}
	}
	return errors.Join(problems...)
}

func (a *ModuleEnv) ApplySubmodule(fsNamespace billy.Filesystem, configRoot string) error {
	module := NewModuleEnv(fsNamespace, configRoot, a.target, a.targetWorkingDir, nil)
	return module.ApplyConfig()
}

func (a *ModuleEnv) ApplyConfig() error {
	moduleFiles, err := a.source.ReadDir(a.sourceWorkingDir)
	if err != nil {
		return err
	}
	for _, moduleFile := range moduleFiles {
		if moduleFile.IsDir() || !strings.HasSuffix(moduleFile.Name(), ".hcl") {
			continue
		}
		if err := (func() error {
			handle, err := a.source.Open(a.source.Join(a.sourceWorkingDir, moduleFile.Name()))
			if err != nil {
				return err
			}
			defer handle.Close()
			contents, err := io.ReadAll(handle)
			if err != nil {
				return err
			}

			cfg := &File{}
			if err := hclsimple.Decode(moduleFile.Name(), contents, nil, cfg); err != nil {
				return err
			}
			return a.resolve(cfg)
		})(); err != nil {
			return err
		}
	}
	return nil
}

type ResourceError struct {
	Type    string
	Name    string
	Problem error
}

func (r *ResourceError) Error() string {
	return fmt.Sprintf("%s(%q) failed to apply because %s", r.Type, r.Name, r.Problem.Error())
}
