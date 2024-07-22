package main

import (
	"context"
	"fmt"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/meschbach/empire/service/resources"
)

type app struct {
	base   string
	fsRoot string
	boxed  string
}

func (a *app) Serve(ctx context.Context) error {
	fmt.Printf("Staring Empire( root %q, base %q, boxed %q)\n", a.fsRoot, a.base, a.boxed)
	local := osfs.New(a.fsRoot)
	var outputFileSystem billy.Filesystem
	if len(a.boxed) > 0 {
		if boxed, err := local.Chroot(a.boxed); err != nil {
			fmt.Printf("Failed to set root because %s", err.Error())
			return nil
		} else {
			outputFileSystem = boxed
		}
	} else {
		outputFileSystem = local
	}

	rootModule := resources.NewModuleEnv(local, a.base, outputFileSystem, "/", &resources.CapturingPackageManager{})
	if err := rootModule.ApplyConfig(); err != nil {
		fmt.Printf("Failed to apply root module because %s\n", err.Error())
	}
	return nil
}
