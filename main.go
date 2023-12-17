package main

import (
	"net/netip"
	"os"
	"time"

	"go.bbkane.com/shovel/digcombine"
	"go.bbkane.com/shovel/serve"
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
		"Dig combinations of QNames/RTypes/Subnets/NSs and summarize results",
		digcombine.Run,
		command.Footer(digFooter),
		command.Flag(
			"--count",
			"Number of times to dig",
			scalar.Int(
				scalar.Default(1),
			),
			flag.ConfigPath("dig.count"),
			flag.Required(),
			flag.Alias("-c"),
		),
		command.Flag(
			"--qname",
			"Qualified names to dig",
			slice.String(),
			flag.ConfigPath("dig.qnames"),
			flag.Required(),
			flag.Alias("-q"),
		),
		command.Flag(
			"--rtype",
			"Record types",
			slice.String(
				slice.Default([]string{"A"}),
				slice.Choices("A", "AAAA", "CNAME", "MX", "NS", "TXT"),
			),
			flag.ConfigPath("dig.rtypes"),
			flag.Required(),
			flag.Alias("-r"),
		),
		command.Flag(
			"--nameserver",
			"Nameserver IP + port to query. Example: 198.51.45.9:53 or dns.google:53 . Set to 'all' to use everything in --nameserver-map",
			slice.String(),
			flag.ConfigPath("dig.nameservers"),
			flag.Required(),
			flag.Alias("-n"),
			flag.UnsetSentinel("UNSET"),
		),
		command.Flag(
			"--nameserver-map",
			"Map of name to nameserver IP:port. Can then use names as arguments to --nameserver",
			dict.String(),
			flag.ConfigPath("dig.nameserver-map"),
		),
		command.Flag(
			"--subnet",
			"Optional client subnet. Example: 101.251.8.0 for China. Set to 'all' to use everything in --subnet-map",
			slice.String(),
			flag.ConfigPath("dig.subnets"),
			flag.Alias("-s"),
			flag.UnsetSentinel("UNSET"),
		),
		command.Flag(
			"--subnet-map",
			"Map of name to subnet. Can then use names as arguments to --subnet",
			dict.Addr(),
			flag.ConfigPath("dig.subnet-map"),
		),
		command.Flag(
			"--global-timeout",
			"Timeout for combined DNS requests",
			scalar.Duration(
				scalar.Default(30*time.Second),
			),
			flag.Required(),
			flag.ConfigPath("dig.global-timeout"),
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
			flag.ConfigPath("dig.protocol"),
		),
	)
}

func serveCmd(digFooter string) command.Command {
	return command.New(
		"Run dig commands remotely",
		serve.Run,
		command.Footer(digFooter),
		command.Flag(
			"--addr-port",
			"Address + Port to serve from",
			scalar.AddrPort(
				scalar.Default(netip.MustParseAddrPort("127.0.0.1:8080")),
			),
			flag.Required(),
			flag.ConfigPath("serve.addr-port"),
		),
		command.Flag(
			"--http-origin",
			"HTTP Origin clients will access",
			scalar.String(
				scalar.Default("http://127.0.0.1:8080"),
			),
			flag.Required(),
			flag.ConfigPath("serve.http-origin"),
		),
		command.Flag(
			"--motd",
			"Message of the day to print on /",
			scalar.String(),
			flag.ConfigPath("serve.motd"),
		),
		command.Flag(
			"--otel-provider",
			"Enable tracing with OpenObserve",
			scalar.String(
				scalar.Choices("open-observe", "stdout"),
				scalar.Default("stdout"),
			),
			flag.Required(),
			flag.ConfigPath("serve.otel.provider"),
		),
		command.Flag(
			"--open-observe-endpoint",
			"Endpoint to send traces to",
			scalar.AddrPort(
				scalar.Default(netip.MustParseAddrPort("127.0.0.1:5080")),
			),
			flag.ConfigPath("serve.open-observe.endpoint"),
		),
		command.Flag(
			"--open-observe-user",
			"OpenObserve Username",
			scalar.String(),
			flag.ConfigPath("serve.open-observe.user"),
			flag.EnvVars("SHOVEL_SERVE_OPENOBSERVE_USER"),
		),
		command.Flag(
			"--open-observe-pass",
			"OpenObserve Password",
			scalar.String(),
			flag.ConfigPath("serve.open-observe.pass"),
			flag.EnvVars("SHOVEL_SERVE_OPENOBSERVE_PASS"),
		),
		command.Flag(
			"--otel-service-name",
			"OpenObserve Service Name",
			scalar.String(
				scalar.Default("shovel serve"),
			),
			flag.ConfigPath("serve.otel.service-name"),
			flag.Required(),
		),
		command.Flag(
			"--otel-service-version",
			"OpenObserve Service Version - TODO: read from shovel binary",
			scalar.String(
				scalar.Default("TODO"),
			),
			flag.ConfigPath("serve.otel.service-version"),
			flag.Required(),
		),
		command.Flag(
			"--otel-service-env",
			"OpenObserve Service Environment",
			scalar.String(
				scalar.Default("dev"),
			),
			flag.ConfigPath("serve.otel.service-env"),
			flag.Required(),
		),
	)
}

func buildApp() warg.App {
	digFooter := `Homepage: https://github.com/bbkane/shovel
Examples: https://github.com/bbkane/shovel/blob/master/examples.md
`

	app := warg.New(
		"shovel",
		section.New(
			"Query DNS and count results. Parameters are combined",
			section.ExistingCommand(
				"dig",
				digCombineCmd(digFooter),
			),
			section.ExistingCommand(
				"serve",
				serveCmd(digFooter),
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
