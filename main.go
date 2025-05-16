package main

import (
	"net/netip"
	"time"

	"go.bbkane.com/shovel/digcombine"
	"go.bbkane.com/shovel/diglist"
	"go.bbkane.com/shovel/serve"
	"go.bbkane.com/warg"
	"go.bbkane.com/warg/command"
	"go.bbkane.com/warg/config/yamlreader"
	"go.bbkane.com/warg/flag"
	"go.bbkane.com/warg/path"
	"go.bbkane.com/warg/section"
	"go.bbkane.com/warg/value/dict"
	"go.bbkane.com/warg/value/scalar"
	"go.bbkane.com/warg/value/slice"
	"go.bbkane.com/warg/wargcore"
)

var version string

func digCombineCmd(digFooter string) wargcore.Command {
	return command.New(
		"Dig combinations of QNames/RTypes/Subnets/NSs and summarize results",
		digcombine.Run,
		command.Footer(digFooter),
		command.NewFlag(
			"--count",
			"Number of times to dig",
			scalar.Int(
				scalar.Default(1),
			),
			flag.ConfigPath("dig.combine.count"),
			flag.Required(),
			flag.Alias("-c"),
		),
		command.NewFlag(
			"--qname",
			"Qualified names to dig",
			slice.String(),
			flag.ConfigPath("dig.combine.qnames"),
			flag.Required(),
			flag.Alias("-q"),
		),
		command.NewFlag(
			"--rtype",
			"Record types",
			slice.String(
				slice.Default([]string{"A"}),
				slice.Choices("A", "AAAA", "CNAME", "MX", "NS", "TXT"),
			),
			flag.ConfigPath("dig.combine.rtypes"),
			flag.Required(),
			flag.Alias("-r"),
		),
		command.NewFlag(
			"--nameserver",
			"Nameserver IP + port to query. Example: 198.51.45.9:53 or dns.google:53 . Set to 'all' to use everything in --nameserver-map",
			slice.String(),
			flag.ConfigPath("dig.combine.nameservers"),
			flag.Required(),
			flag.Alias("-n"),
			flag.UnsetSentinel("UNSET"),
		),
		command.NewFlag(
			"--nameserver-map",
			"Map of name to nameserver IP:port. Can then use names as arguments to --nameserver",
			dict.String(),
			flag.ConfigPath("dig.combine.nameserver-map"),
		),
		command.NewFlag(
			"--subnet",
			"Optional client subnet. Example: 101.251.8.0 for China. Set to 'all' to use everything in --subnet-map",
			slice.String(),
			flag.ConfigPath("dig.combine.subnets"),
			flag.Alias("-s"),
			flag.UnsetSentinel("UNSET"),
		),
		command.NewFlag(
			"--subnet-map",
			"Map of name to subnet. Can then use names as arguments to --subnet",
			dict.Addr(),
			flag.ConfigPath("dig.combine.subnet-map"),
		),
		command.NewFlag(
			"--global-timeout",
			"Timeout for combined DNS requests",
			scalar.Duration(
				scalar.Default(30*time.Second),
			),
			flag.Required(),
			flag.ConfigPath("dig.combine.global-timeout"),
		),
		command.NewFlag(
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

func digListCmd(digFooter string) wargcore.Command {
	return command.New(
		"Pruduces digs from a list of inputs and prints the summarized results as YAML",
		diglist.Run,
		command.Footer(digFooter),
		command.NewFlag(
			"--count",
			"Number of times to dig",
			slice.Int(),
			flag.ConfigPath("dig.list[].count"),
			flag.Required(),
			flag.Alias("-c"),
		),
		command.NewFlag(
			"--qname",
			"qualified names to dig",
			slice.String(),
			flag.ConfigPath("dig.list[].qname"),
			flag.Required(),
			flag.Alias("-q"),
		),
		command.NewFlag(
			"--nameserver",
			"Nameserver IP + port to query. Example: 198.51.45.9:53 or dns.google:53",
			slice.String(),
			flag.ConfigPath("dig.list[].nameserver"),
			flag.Required(),
			flag.Alias("-n"),
			flag.UnsetSentinel("UNSET"),
		),
		command.NewFlag(
			"--protocol",
			"Protocol to use when digging",
			slice.String(
				slice.Choices("udp", "udp4", "udp6", "tcp", "tcp4", "tcp6"),
			),
			flag.Required(),
			flag.Alias("-p"),
			flag.ConfigPath("dig.list[].protocol"),
		),
		command.NewFlag(
			"--rtype",
			"Record types",
			slice.String(
				slice.Choices("A", "AAAA", "CNAME", "MX", "NS", "TXT"),
			),
			flag.ConfigPath("dig.list[].rtype"),
			flag.Required(),
			flag.Alias("-r"),
		),
		command.NewFlag(
			"--subnet",
			"Client subnet. Example: 101.251.8.0 for China. Set to 'none' not use a client subnet",
			slice.String(),
			flag.ConfigPath("dig.list[].subnet"),
			flag.Alias("-s"),
			flag.Required(),
			flag.UnsetSentinel("UNSET"),
		),
		command.NewFlag(
			"--timeout",
			"Timeout for each individual DNS request",
			slice.Duration(),
			flag.Required(),
			flag.ConfigPath("dig.list[].timeout"),
		),
	)
}

func serveCmd(digFooter string) wargcore.Command {
	return command.New(
		"Run dig commands remotely",
		serve.Run,
		command.Footer(digFooter),
		command.NewFlag(
			"--addr-port",
			"Address + Port to serve from",
			scalar.AddrPort(
				scalar.Default(netip.MustParseAddrPort("127.0.0.1:8080")),
			),
			flag.Required(),
			flag.ConfigPath("serve.addr-port"),
		),
		command.NewFlag(
			"--footer",
			"Trailing HTML for the bottom of the page",
			scalar.String(
				scalar.Default(""),
			),
			flag.ConfigPath("serve.footer"),
		),
		command.NewFlag(
			"--https-certfile",
			"Path to HTTP public key in PEM format. NOTE: in most cases, this is the leaf cert concatenated with the ICA in one file, and the order matters",
			scalar.Path(),
			flag.ConfigPath("serve.https.certfile"),
		),
		command.NewFlag(
			"--https-keyfile",
			"Path to HTTP private key in PEM format",
			scalar.Path(),
			flag.ConfigPath("serve.https.keyfile"),
		),
		command.NewFlag(
			"--motd",
			"Message of the day to print on / . Should be HTML",
			scalar.String(
				scalar.Default("<p style=\"text-align:center\">Send and summarize multiple DNS queries</p>"),
			),
			flag.ConfigPath("serve.motd"),
		),
		command.NewFlag(
			"--otel-provider",
			"Enable tracing with OpenObserve",
			scalar.String(
				scalar.Choices("openobserve", "stdout"),
				scalar.Default("stdout"),
			),
			flag.EnvVars("SHOVEL_SERVE_OTEL_PROVIDER"),
			flag.Required(),
			flag.ConfigPath("serve.otel.provider"),
		),
		command.NewFlag(
			"--openobserve-endpoint",
			"Endpoint to send traces to",
			scalar.AddrPort(
				scalar.Default(netip.MustParseAddrPort("127.0.0.1:5080")),
			),
			flag.ConfigPath("serve.openobserve.endpoint"),
		),
		command.NewFlag(
			"--openobserve-user",
			"OpenObserve Username",
			scalar.String(),
			flag.ConfigPath("serve.openobserve.user"),
			flag.EnvVars("SHOVEL_SERVE_OPENOBSERVE_USER"),
		),
		command.NewFlag(
			"--openobserve-pass",
			"OpenObserve Password",
			scalar.String(),
			flag.ConfigPath("serve.openobserve.pass"),
			flag.EnvVars("SHOVEL_SERVE_OPENOBSERVE_PASS"),
		),
		command.NewFlag(
			"--otel-service-env",
			"OpenObserve Service Environment",
			scalar.String(
				scalar.Default("dev"),
			),
			flag.ConfigPath("serve.otel.service-env"),
			flag.Required(),
		),
		command.NewFlag(
			"--protocol",
			"Serve via HTTP or HTTPS",
			scalar.String(
				scalar.Choices("HTTP", "HTTPS"),
				scalar.Default("HTTP"),
			),
			flag.ConfigPath("serve.protocol"),
			flag.EnvVars("SHOVEL_SERVE_PROTOCOL"),
			flag.Required(),
		),
		command.NewFlag(
			"--trace-id-template",
			"Go HTML Template to customize TraceID formatting. Available fields: .TraceID",
			scalar.String(
				scalar.Default("<p>TraceID: <a href=\"http://localhost:5080/web/traces?period=15m&query=&org_identifier=default&trace_id={{.TraceID}}\" target=\"_blank\">{{.TraceID}}</a></p>"),
			),
			flag.ConfigPath("serve.trace-id-template"),
			flag.Required(),
		),
	)
}

func buildApp() *wargcore.App {
	digFooter := `Homepage: https://github.com/bbkane/shovel
Examples: https://github.com/bbkane/shovel/blob/master/examples.md
`

	app := warg.New(
		"shovel",
		version,
		section.New(
			"Query DNS and count results",
			section.Command(
				"serve",
				serveCmd(digFooter),
			),
			section.CommandMap(warg.VersionCommandMap()),
			section.Footer(digFooter),
			section.NewSection(
				"dig",
				"Dig in different ways",
				section.Command(
					"combine",
					digCombineCmd(digFooter),
				),
				section.Command(
					"list",
					digListCmd(digFooter),
				),
			),
		),
		warg.ConfigFlag(
			yamlreader.New,
			wargcore.FlagMap{
				"--config": flag.New(
					"Path to YAML config file",
					scalar.Path(
						scalar.Default(path.New("~/.config/grabbit.yaml")),
					),
				),
			},
		),
		warg.GlobalFlagMap(warg.ColorFlagMap()),
	)
	return &app
}

func main() {
	app := buildApp()
	app.MustRun()
}
