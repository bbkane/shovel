package serve

import (
	"fmt"
	"strings"

	"go.bbkane.com/shovel/dig"
	"gopkg.in/yaml.v3"
)

// buildTableJSON returns a YAML string suitable so we can copy it from the button. Copied from digList for now... will probably want to improve the format later.
func buildTableYAML(dRes []dig.DigRepeatResult) (string, error) {
	type Rdata struct {
		Content []string `yaml:"content"`
		Count   int      `yaml:"count"`
	}

	type Error struct {
		Count int    `yaml:"count"`
		Msg   string `yaml:"msg"`
	}

	type Result struct {
		Rdata  []Rdata `yaml:"rdata"`
		Errors []Error `yaml:"errors"`
	}

	type Return struct {
		Results []Result `yaml:"results"`
	}

	ret := Return{
		Results: make([]Result, len(dRes)),
	}
	for i := range dRes {
		ret.Results[i].Rdata = make([]Rdata, len(dRes[i].Answers))
		for r := range dRes[i].Answers {
			ret.Results[i].Rdata[r].Content = dRes[i].Answers[r].StringSlice
			ret.Results[i].Rdata[r].Count = dRes[i].Answers[r].Count
		}

		ret.Results[i].Errors = make([]Error, len(dRes[i].Errors))
		for e := range dRes[i].Errors {
			ret.Results[i].Errors[e].Msg = dRes[i].Errors[e].String
			ret.Results[i].Errors[e].Count = dRes[i].Errors[e].Count

		}
	}

	b := strings.Builder{}
	encoder := yaml.NewEncoder(&b)
	encoder.SetIndent(2)
	err := encoder.Encode(&ret)
	if err != nil {
		return "", fmt.Errorf("could not serialize digResult to yaml: %w", err)
	}
	return b.String(), nil
}
