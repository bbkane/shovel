package main

import (
	"os"
	"time"

	"go.bbkane.com/warg"
	"go.bbkane.com/warg/command"
	"go.bbkane.com/warg/flag"
	"go.bbkane.com/warg/section"
	"go.bbkane.com/warg/value/scalar"
)

var version string

func buildApp() warg.App {
	app := warg.New(
		"shovel",
		section.New(
			"Dig some stuff!",
			section.Command(
				"dig",
				"simple dig",
				runDig,
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
					"--nameserver-ip",
					"Nameserver to query",
					scalar.String(
						scalar.Default("198.51.45.9"), // dns2.p09.nsone.net.
					),
					flag.Required(),
				),
				command.Flag(
					"--nameserver-port",
					"Port on the nameserver",
					scalar.Int(
						scalar.Default(53),
					),
					flag.Required(),
				),
				command.Flag(
					"--subnet-ip",
					"Optional client subnet. 101.251.8.0 for China for example",
					scalar.String(),
				),
				command.Flag(
					"--timeout",
					"Timeout for request",
					scalar.Duration(
						scalar.Default(2*time.Second),
					),
					flag.Required(),
				),
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
