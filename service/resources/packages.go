package resources

import (
	"errors"
	"fmt"
)

type Package struct {
	Name string `hcl:"name,label"`
}

func (p *Package) Perform(env *ModuleEnv) error {
	fmt.Printf("Install package %q\n", p.Name)
	if err := env.Runtime.EnsureInstalled(p.Name); err != nil {
		return err
	}

	if !env.Completed("package", p.Name) {
		return errors.New("failed to register resource as complete")
	}
	return nil
}
