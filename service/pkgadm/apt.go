package pkgadm

import (
	"context"
	"errors"
)

type PackageSpec struct {
	Name string
}

type ExecResult struct {
	Code   int
	Stdout string
	Stderr string
}

type Exec interface {
	Run(ctx context.Context, cmd []string) (ExecResult, error)
}

type Apt struct {
}

func (a *Apt) Install(ctx context.Context, env Exec, spec *PackageSpec) error {
	result, err := env.Run(ctx, []string{"apt-get", "install", "-y", spec.Name})
	if err != nil {
		return err
	}
	if result.Code != 0 {
		return errors.New(result.Stdout)
	}
	return nil
}

func (a *Apt) Update(ctx context.Context, env Exec) error {
	result, err := env.Run(ctx, []string{"apt-get", "update"})
	if err != nil {
		return err
	}
	if result.Code != 0 {
		return errors.New(result.Stdout)
	}
	return nil
}

func (a *Apt) Hold(ctx context.Context, env Exec) error {
	result, err := env.Run(ctx, []string{"apt-mark", "hold"})
	if err != nil {
		return err
	}
	if result.Code != 0 {
		return errors.New(result.Stdout)
	}
	return nil
}
