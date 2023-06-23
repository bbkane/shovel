package main

import (
	"os"
	"time"

	"go.bbkane.com/shovel/digcombine"
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

func digCombineCmd(digFooter string) command.Command {
	return command.New(
		"Dig combinations of FQDNs/RTypes/Subnets/NSs and summarize results",
		digcombine.Run,
		command.Footer(digFooter),
		command.Flag(
			"--count",
			"Number of times to dig",
			scalar.Int(
				scalar.Default(1),
			),
			flag.ConfigPath("dig.combine.count"),
			flag.Required(),
			flag.Alias("-c"),
		),
		command.Flag(
			"--fqdn",
			"FQDNs to dig",
			slice.String(),
			flag.ConfigPath("dig.combine.fqdns"),
			flag.Required(),
			flag.Alias("-f"),
		),
		command.Flag(
			"--rtype",
			"Record types",
			slice.String(
				slice.Default([]string{"A"}),
				slice.Choices("A", "AAAA", "CNAME", "TXT"),
			),
			flag.ConfigPath("dig.combine.rtypes"),
			flag.Required(),
			flag.Alias("-r"),
		),
		command.Flag(
			"--ns",
			"Nameserver IP + port to query. Example: 198.51.45.9:53 or dns.google:53 . Set to 'all' to use everything in --ns-map",
			slice.String(),
			flag.ConfigPath("dig.combine.nameservers"),
			flag.Required(),
			flag.Alias("-n"),
		),
		command.Flag(
			"--ns-map",
			"Map of name to nameserver IP:port. Can then use names as arguments to --ns",
			dict.String(),
			flag.ConfigPath("dig.combine.nameserver-map"),
		),
		command.Flag(
			"--subnet",
			"Optional client subnet. Example: 101.251.8.0 for China. Set to 'all' to use everything in --subnet-map",
			slice.String(),
			flag.ConfigPath("dig.combine.subnets"),
			flag.Alias("-s"),
		),
		command.Flag(
			"--subnet-map",
			"Map of name to subnet. Can then use names as arguments to --subnet",
			dict.Addr(),
			flag.ConfigPath("dig.combine.subnet-map"),
		),
		command.Flag(
			"--timeout",
			"Timeout for each individual DNS request",
			scalar.Duration(
				scalar.Default(2*time.Second),
			),
			flag.Required(),
			flag.ConfigPath("dig.combine.timeout"),
		),
		command.Flag(
			"--mock-dig-func",
			"Flag to mock dig func. Used only for testing",
			scalar.String(
				scalar.Default("none"),
			),
			flag.Required(),
		),
		command.Flag(
			"--protocol",
			"Protocol to use when digging",
			scalar.String(
				scalar.Choices("udp", "udp4", "udp6", "tcp", "tcp4", "tcp6"),
				scalar.Default("udp"),
			),
			flag.Required(),
			flag.Alias("-p"),
			flag.ConfigPath("dig.combine.protocol"),
		),
	)
}

// func digListCmd()

func buildApp() warg.App {
	digFooter := `Homepage: https://github.com/bbkane/shovel
Examples: https://github.com/bbkane/shovel/blob/master/examples.md
`

	app := warg.New(
		"shovel",
		section.New(
			"Query DNS and count results",
			section.Section(
				"dig",
				"Dig in different ways",
				section.ExistingCommand(
					"combine",
					digCombineCmd(digFooter),
				),
			),
			section.Footer(digFooter),
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
