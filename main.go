package main

import (
	"os"
	"time"

	"go.bbkane.com/warg"
	"go.bbkane.com/warg/command"
	"go.bbkane.com/warg/config/yamlreader"
	"go.bbkane.com/warg/flag"
	"go.bbkane.com/warg/section"
	"go.bbkane.com/warg/value/dict"
	"go.bbkane.com/warg/value/scalar"
	"go.bbkane.com/warg/value/slice"
)

var version string

func digCmd() command.Command {
	return command.New(
		"Query DNS and count results",
		runDig,
		command.Flag(
			"--count",
			"Number of times to dig",
			scalar.Int(
				scalar.Default(10),
			),
			flag.ConfigPath("dig.count"),
			flag.Required(),
		),
		command.Flag(
			"--fqdn",
			"FQDNs to dig",
			slice.String(),
			flag.ConfigPath("dig.fqdns"),
			flag.Required(),
			flag.Alias("-f"),
		),
		command.Flag(
			"--rtype",
			"Record type",
			slice.String(
				slice.Default([]string{"A"}),
				slice.Choices("A", "AAAA", "CNAME"),
			),
			flag.ConfigPath("dig.rtypes"),
			flag.Required(),
			flag.Alias("-r"),
		),
		command.Flag(
			"--ns",
			"Nameserver IP + port to query. Example: 198.51.45.9:53",
			slice.String(),
			flag.ConfigPath("dig.nameservers"),
			flag.Required(),
		),
		command.Flag(
			"--ns-map",
			"Map of name to nameserver IP. Can then use names as arguments to --ns",
			dict.AddrPort(),
			flag.ConfigPath("dig.nameserver-map"),
		),
		command.Flag(
			"--subnet",
			"Optional client subnet. 101.251.8.0 for China for example",
			slice.String(),
			flag.ConfigPath("dig.subnets"),
			flag.Alias("-s"),
		),
		command.Flag(
			"--subnet-map",
			"Map of name to subnet. Can then use names as arguments to --subnet",
			dict.Addr(),
			flag.ConfigPath("dig.subnet-map"),
		),
		command.Flag(
			"--timeout",
			"Timeout for each individual DNS request",
			scalar.Duration(
				scalar.Default(2*time.Second),
			),
			flag.Required(),
			flag.ConfigPath("dig.timeout"),
		),
	)
}

func buildApp() warg.App {
	app := warg.New(
		"shovel",
		section.New(
			"Query DNS and count results",
			section.ExistingCommand(
				"dig",
				digCmd(),
			),
		),
		warg.AddColorFlag(),
		warg.AddVersionCommand(version),
		warg.ConfigFlag(
			"--config",
			[]scalar.ScalarOpt[string]{
				scalar.Default("~/.config/shovel.yaml"),
			},
			yamlreader.New,
			"Config",
		),
	)
	return app
}

func main() {
	app := buildApp()
	// app.MustRun([]string{"shovel", "dig", "--rtype", "A"}, os.LookupEnv)
	app.MustRun(os.Args, os.LookupEnv)
}
