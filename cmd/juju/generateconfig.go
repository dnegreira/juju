package main

import (
	"fmt"
	"launchpad.net/gnuflag"
	"launchpad.net/juju-core/cmd"
	"launchpad.net/juju-core/environs"
	"os"
)

// GenerateConfigCommand is used to write out a boilerplate environments.yaml file.
type GenerateConfigCommand struct {
	WriteFile bool
}

func (c *GenerateConfigCommand) Info() *cmd.Info {
	return &cmd.Info{"generate-config", "", "generate boilerplate configuration for juju environments", ""}
}

func (c *GenerateConfigCommand) Init(f *gnuflag.FlagSet, args []string) error {
	f.BoolVar(&c.WriteFile, "w", false, "write to environments.yaml file if it doesn't already exist")
	if err := f.Parse(true, args); err != nil {
		return err
	}
	return cmd.CheckEmpty(f.Args())
}

// Run checks to see if there is already an environments.yaml file. In one does not exist already,
// a boilerplate version is created so that the user can edit it to get started.
func (c *GenerateConfigCommand) Run(context *cmd.Context) error {
	out := context.Stdout
	config := environs.BoilerplateConfig()
	if !c.WriteFile {
		fmt.Fprintln(out, config)
		return nil
	}
	_, err := environs.ReadEnvirons("")
	if err == nil {
		fmt.Fprintf(out, "A juju environment configuration already exists.\n")
		fmt.Fprintf(out, "It will not be overwritten.\n")
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}
	filename, err := environs.WriteEnvirons("", config)
	if err != nil {
		return fmt.Errorf("A boilerplate environment configuration file could not be created: %s", err.Error())
	}
	fmt.Fprintf(out, "A boilerplate environment configuration file has been written to %s.\n", filename)
	fmt.Fprint(out, "Edit the file to configure your juju environment and re-run bootstrap.\n")
	return nil
}
