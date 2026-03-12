package main

import (
	"fmt"

	"github.com/myselfBZ/trailpkg/internal/manifest"
)

type app struct {
	input           UserInput
	manifestManager *manifest.ManifestManager
}

type Verb string

const (
	Install Verb = "install"
	Help    Verb = "help"
	Update  Verb = "update"
	Remove  Verb = "rm"
)

type UserInput struct {
	Verb Verb
	VerbArgs []string
}

func newUserInput(args []string) UserInput {
	input := UserInput{
		Verb: Verb(args[1]),
	}

	if len(args) >= 3 {
		input.VerbArgs = append(input.VerbArgs, args[2:]...)
	}

	return input
}

func (a *app) executeUserInput() ExitStatus {
	exitStatus := ExitStatus{
		Code: StatusSuccess,
	}

	switch a.input.Verb {

	case Install:
		if len(a.input.VerbArgs) != 1 {
			exitStatus.Code = StatusWrongNumberArgs
			exitStatus.Message = "'install' requires only 1 argument"
			break
		}

		if err := a.handleInstall(a.input.VerbArgs[0]); err != nil {
			exitStatus.Code = StatusInstallationFailed
			exitStatus.Message = err.Error()
		}
	case Remove:
		if len(a.input.VerbArgs) != 1 {
			exitStatus.Code = StatusWrongNumberArgs
			exitStatus.Message = "'install' requires only 1 argument"
			break
		}

		if err := a.handleRemove(a.input.VerbArgs[0]); err != nil {
			exitStatus.Code = StatusRemovalFailed
			exitStatus.Message = err.Error()
			break
		}
		fmt.Printf("Package %s has been removed successfully\n", a.input.VerbArgs[0])

	case Update:
		if len(a.input.VerbArgs) > 0 {
			exitStatus.Code = StatusWrongNumberArgs
			exitStatus.Message = "'update' does not require any arguments"
			break
		}

		a.handleUpdate()

	case Help:
		if len(a.input.VerbArgs) > 0 {
			exitStatus.Code = StatusWrongNumberArgs
			exitStatus.Message = "'help' does not take any arguments"
			break
		}

		a.handleHelp()
	}

	return exitStatus

}

func (a *app) handleInstall(pkgName string) error {
	fmt.Printf("Installing %s...\n", pkgName)
	if err := a.manifestManager.Install(pkgName); err != nil {
		return err
	}
	return nil
}


func (a *app) handleUpdate() {
	if err := a.manifestManager.UpdateManifest(); err != nil {
		fmt.Println("update error: ", err)
	}
}

func (a *app) handleHelp() {
	fmt.Println("Version: 0.0.0.0.1")
	fmt.Println("This is trailpkg")
	fmt.Println("To install packages, run 'trail install <package>'")
}

func (a *app) handleRemove(pkgName string) error {
	fmt.Printf("Removing %s...\n", pkgName)
	if err := a.manifestManager.Remove(pkgName); err != nil {
		return err
	}
	return nil
}
