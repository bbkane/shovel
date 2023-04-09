package main

import (
	"os"
	"time"

	"go.bbkane.com/warg"
	"go.bbkane.com/warg/command"
	"go.bbkane.com/warg/flag"
	"go.bbkane.com/warg/section"
	"go.bbkane.com/warg/value/dict"
	"go.bbkane.com/warg/value/scalar"
)

var version string

func digOneCommand() command.Command {
	return command.New(
		"Query DNS and count results",
		runDig,
		command.Flag(
			"--count",
			"Number of times to dig",
			scalar.Int(
				scalar.Default(1),
			),
			flag.Required(),
		),
		command.Flag(
			"--fqdn",
			"FQDN to dig",
			scalar.String(
				scalar.Default("linkedin.com"),
			),
			flag.Required(),
		),
		command.Flag(
			"--rtype",
			"Record type",
			scalar.String(
				scalar.Default("A"),
				scalar.Choices("A", "AAAA", "CNAME"),
			),
			flag.Required(),
		),
		command.Flag(
			"--ns",
			"Nameserver IP + port to query. Example: 198.51.45.9:53",
			scalar.String(
				scalar.Default("198.51.45.9:53"),
			),
			flag.Required(),
		),
		command.Flag(
			"--ns-map",
			"Map of name to nameserver IP. Can then use names as arguments to --ns",
			dict.AddrPort(),
		),
		command.Flag(
			"--subnet",
			"Optional client subnet. 101.251.8.0 for China for example",
			scalar.String(),
		),
		command.Flag(
			"--subnet-map",
			"Map of name to subnet. Can then use names as arguments to --subnet",
			dict.Addr(),
		),
		command.Flag(
			"--timeout",
			"Timeout for request",
			scalar.Duration(
				scalar.Default(2*time.Second),
			),
			flag.Required(),
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
				digOneCommand(),
			),
		),
		warg.AddColorFlag(),
		warg.AddVersionCommand(version),
	)
	return app
}

func main() {
	app := buildApp()
	// app.MustRun([]string{"shovel", "dig", "--rtype", "A"}, os.LookupEnv)
	app.MustRun(os.Args, os.LookupEnv)
}
