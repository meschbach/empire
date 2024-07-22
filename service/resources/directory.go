package resources

type Directory struct {
	Name string `hcl:"name,label"`
	Path string `hcl:"path"`
}

func (d *Directory) Perform(env *ModuleEnv) error {
	targetPath, targetPathError := env.ResolveAbsolutePath(d.Path)
	if targetPathError != nil {
		return targetPathError
	}
	if exists, err := env.EnsureFileExists(targetPath); err != nil {
		return err
	} else if exists {
		env.Completed("directory", d.Name)
	} else {
		if err := env.MakeDirectory(targetPath); err != nil {
			return err
		}
		env.Completed("directory", d.Name)
	}
	return nil
}
