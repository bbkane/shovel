package main

import (
	"os"
	"testing"

	"go.bbkane.com/warg"
)

func TestBuildApp(t *testing.T) {
	app := buildApp()

	if err := app.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestRunCLI(t *testing.T) {
	updateGolden := os.Getenv("SHOVEL_TEST_UPDATE_GOLDEN") != ""
	tests := []struct {
		name   string
		app    warg.App
		args   []string
		lookup warg.LookupFunc
	}{
		{
			name: "simple",
			app:  buildApp(),
			args: []string{"shovel", "dig", "combine",
				"--config", "notthere", // Hack so shovel doesn't try to read a config
				"--count", "1",
				"--qname", "linkedin.com",
				"--mock-dig-func", "simple", // don't really dig!
				"--ns", "0.0.0.0:53",
				"--rtype", "A",
			},
			lookup: warg.LookupMap(nil),
		},
		{
			name: "twocount",
			app:  buildApp(),
			args: []string{"shovel", "dig", "combine",
				"--config", "notthere", // Hack so shovel doesn't try to read a config
				"--count", "2",
				"--qname", "linkedin.com",
				"--mock-dig-func", "twocount", // don't really dig!
				"--ns", "0.0.0.0:53",
				"--rtype", "A",
			},
			lookup: warg.LookupMap(nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warg.GoldenTest(t, tt.app, tt.args, tt.lookup, updateGolden)
		})
	}
}
