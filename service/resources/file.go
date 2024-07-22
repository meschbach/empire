package resources

import (
	"github.com/hashicorp/hcl/v2/hclsimple"
)

type File struct {
	Directory        []Directory       `hcl:"directory,block"`
	StaticLocalFiles []StaticLocalFile `hcl:"static_local_file,block"`
	GitModule        []GitModule       `hcl:"git_module,block"`
	Package          []Package         `hcl:"package,block"`
}

func Parse(fileName string) (*File, error) {
	var config File
	if err := hclsimple.DecodeFile(fileName, nil, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
