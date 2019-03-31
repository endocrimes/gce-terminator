package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/endocrimes/gce-terminator/reaper"
	"github.com/genuinetools/pkg/cli"
	hclog "github.com/hashicorp/go-hclog"
)

var (
	googleCloudProject       string
	googleCloudZone          string
	googleCloudInstanceGroup string

	interval string

	logger hclog.Logger
)

type runCommand struct{}

const runHelp = `Run the terminator.`

func (cmd *runCommand) Name() string      { return "run" }
func (cmd *runCommand) Args() string      { return "" }
func (cmd *runCommand) ShortHelp() string { return runHelp }
func (cmd *runCommand) LongHelp() string  { return runHelp }
func (cmd *runCommand) Hidden() bool      { return false }

func (cmd *runCommand) Register(fs *flag.FlagSet) {}

func (cmd *runCommand) Run(ctx context.Context, args []string) error {
	cfg := &reaper.Config{
		GCPProject:        googleCloudProject,
		GCPZone:           googleCloudZone,
		InstanceGroupName: googleCloudInstanceGroup,
	}

	if interval != "" {
		d, err := time.ParseDuration(interval)
		if err != nil {
			return fmt.Errorf("Parsing duration failed: %v", err)
		}

		cfg.PollInterval = &d
	}

	return reaper.NewReaper(cfg, logger).Run(ctx)
}

func main() {
	p := cli.NewProgram()
	p.Name = "buildkite-gcp-scaler"
	p.Description = `A tool that autoscales Google Cloud clusters to run Buildkite jobs`

	// Setup the global flags.
	var (
		debug bool
	)
	p.FlagSet = flag.NewFlagSet("global", flag.ExitOnError)
	p.FlagSet.BoolVar(&debug, "d", false, "enable debug logging")
	p.FlagSet.StringVar(&googleCloudProject, "gcp-project", "", "Google Cloud Project")
	p.FlagSet.StringVar(&googleCloudZone, "gcp-zone", "", "Google Cloud Zone")
	p.FlagSet.StringVar(&googleCloudInstanceGroup, "instance-group", "", "Google Cloud Instance Group")
	p.FlagSet.StringVar(&interval, "interval", "", "How frequently the scaler should run")

	p.Before = func(ctx context.Context) error {
		logLevel := "INFO"
		if debug {
			logLevel = "DEBUG"
		}

		logger = hclog.New(&hclog.LoggerOptions{
			Name:  "gce-terminator",
			Level: hclog.LevelFromString(logLevel),
		})
		return nil
	}

	// Add our commands.
	p.Commands = []cli.Command{
		&runCommand{},
	}

	// Run our program.
	p.Run()
}
