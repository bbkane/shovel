package main

import (
	"net/netip"
	"os"
	"time"

	"go.bbkane.com/warg"
	"go.bbkane.com/warg/command"
	"go.bbkane.com/warg/flag"
	"go.bbkane.com/warg/section"
	"go.bbkane.com/warg/value/scalar"
)

var version string

func digOneCommand() command.Command {
	return command.New(
		"simple dig",
		runDigOne,
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
			"--nameserver-addr-port",
			"Nameserver to query",
			scalar.AddrPort(
				scalar.Default(netip.MustParseAddrPort("198.51.45.9:53")),
			),
			flag.Required(),
		),

		command.Flag(
			"--subnet-addr",
			"Optional client subnet. 101.251.8.0 for China for example",
			scalar.Addr(),
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
			"Dig some stuff!",
			section.ExistingCommand(
				"dig-one",
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
