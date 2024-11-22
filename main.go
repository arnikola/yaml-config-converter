package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	fluentbitconfig "github.com/calyptia/go-fluentbit-config/v2"
)

var usage string = `
FluentBit YAML to INI converter

Usage:
	yaml-config-converter <FILEPATH>
`

func main() {
	if len(os.Args) != 2 {
		panic("Must provide exactly 1 filepath arg\n" + usage)
	}

	ini, err := printConfig(os.Args[1])
	if err != nil {
		panic(err)
	}

	fmt.Println(ini)
}

func printConfig(filepath string) (string, error) {
	var raw []byte
	var err error
	var config Config

	raw, err = os.ReadFile(filepath)
	if err != nil {
		return "", err
	}

	dec := yaml.NewDecoder(strings.NewReader(string(raw)))
	dec.KnownFields(true)
	err = dec.Decode(&config)
	if errors.Is(err, io.EOF) {
		return "", err
	}

	var fbCfg = fluentbitconfig.Config{
		Env:      config.Env,
		Includes: config.Includes,
		Service:  config.Service,
		Customs:  config.Customs,
		Pipeline: config.Pipeline.Pipeline,
	}

	// Gather all parsers across either config format.
	parsers := dedupeParsers(append(fbCfg.Pipeline.Parsers, config.Parsers...))
	fbCfg.Pipeline.Parsers = parsers
	converted, err := fbCfg.DumpAsClassic()
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	// Gather all multiline parsers across either config format.
	multi := dedupeParsers(append(config.Pipeline.Multi, config.Multi...))
	_, err = sb.WriteString(converted)
	if err != nil {
		return "", err
	}

	err = writePlugins(&sb, "MULTILINE_PARSER", multi)
	if err != nil {
		return "", err
	}

	return sb.String(), nil
}
