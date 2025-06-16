package main

import (
	"context"
	"os"
	"testing"

	"go.bbkane.com/shovel/dig"
	"go.bbkane.com/warg"
	"go.bbkane.com/warg/parseopt"
	"go.bbkane.com/warg/wargcore"
)

func TestBuildApp(t *testing.T) {
	t.Parallel()
	app := buildApp()

	if err := app.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestRunCLI(t *testing.T) {
	t.Parallel()
	updateGolden := os.Getenv("SHOVEL_TEST_UPDATE_GOLDEN") != ""
	if !updateGolden {
		t.Log("To update golden files, run: SHOVEL_TEST_UPDATE_GOLDEN=1 go test ./... ")
	}
	tests := []struct {
		name       string
		app        *wargcore.App
		args       []string
		digOneFunc dig.DigOneFunc
	}{
		{
			name: "simple",
			app:  buildApp(),
			args: []string{"shovel", "dig", "combine",
				"--config", "notthere", // Hack so shovel doesn't try to read a config
				"--count", "1",
				"--qname", "linkedin.com",
				"--nameserver", "0.0.0.0:53",
				"--rtype", "A",
			},
			digOneFunc: dig.DigOneFuncMock(
				context.Background(),
				[]dig.DigOneResult{
					{Answers: []string{"1.2.3.4"}, Err: nil},
				},
			),
		},
		{
			name: "twocount",
			app:  buildApp(),
			args: []string{"shovel", "dig", "combine",
				"--config", "notthere", // Hack so shovel doesn't try to read a config
				"--count", "2",
				"--qname", "linkedin.com",
				"--nameserver", "0.0.0.0:53",
				"--rtype", "A",
			},
			digOneFunc: dig.DigOneFuncMock(
				context.Background(),
				[]dig.DigOneResult{
					{Answers: []string{"1.2.3.4"}, Err: nil},
					{Answers: []string{"1.2.3.4"}, Err: nil},
				},
			),
		},
	}

	for _, tt := range tests {
		ctx := context.WithValue(context.Background(), dig.DigOneFuncCtxKey{}, tt.digOneFunc)
		t.Run(tt.name, func(t *testing.T) {
			warg.GoldenTest(
				t,
				warg.GoldenTestArgs{
					App:             tt.app,
					UpdateGolden:    updateGolden,
					ExpectActionErr: false,
				},
				parseopt.Args(tt.args),
				parseopt.Context(ctx),
			)
		})
	}
}
