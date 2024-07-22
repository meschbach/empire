package resources

import (
	"errors"
	"fmt"
)

type StaticLocalFile struct {
	Target string `hcl:"target"`
	Source string `hcl:"source"`

	DependsOn []Dependency `hcl:"depends_on,block"`
}

func (s *StaticLocalFile) Perform(env *ModuleEnv) error {
	allDependenciesDone := true
	for _, dependency := range s.DependsOn {
		if done := env.HasCompleted(dependency.Type, dependency.Name); !done {
			fmt.Printf("Dependency %s(%s) not done.", dependency.Type, dependency.Name)
			allDependenciesDone = false
		}
	}
	if !allDependenciesDone {
		fmt.Printf("Dependencies not done.  Will retry\n")
		return nil //todo: retry logic
	}

	sourcePath, sourcePathErr := env.ResolveConfigFile(s.Source)
	targetPath, targetPathErr := env.ResolveAbsolutePath(s.Target)
	if err := errors.Join(sourcePathErr, targetPathErr); err != nil {
		return err
	}

	if err := env.CopyFile(sourcePath, targetPath, CopyOpt{}); err != nil {
		return err
	}
	env.Completed("static_local_file", s.Target)
	return nil
}
