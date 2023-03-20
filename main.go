package main

import (
	"os"

	"go.bbkane.com/warg"
	"go.bbkane.com/warg/section"
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
				dig,
			),
		),
		warg.AddColorFlag(),
		warg.AddVersionCommand(version),
	)
	return app
}

func main() {
	app := buildApp()
	app.MustRun(os.Args, os.LookupEnv)
}
