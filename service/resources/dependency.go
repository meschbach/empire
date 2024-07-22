package resources

type Dependency struct {
	Type string `hcl:"type"`
	Name string `hcl:"name"`
}
