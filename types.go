package main

import (
	fluentbitconfig "github.com/calyptia/go-fluentbit-config/v2"
	"github.com/calyptia/go-fluentbit-config/v2/property"
)

// Config is taken from https://github.com/chronosphereio/calyptia-go-fluentbit-config/blob/main/config.go#L12
// with the following changes:
//   - `multiline_parsers“ support to allow parsing of newer 3.2 fluentbit syntax
//   - `parsers“ support at top level to allow parsing of newer 3.2 fluentbit syntax
//   - overwriting expected pipeline with WrappedPipe that allows for multiline_parsers support
type Config struct {
	Env      property.Properties     `yaml:"env,omitempty"`
	Includes []string                `yaml:"includes,omitempty"`
	Service  property.Properties     `yaml:"service,omitempty"`
	Customs  fluentbitconfig.Plugins `yaml:"customs,omitempty"`

	Pipeline WrappedPipe             `yaml:"pipeline,omitempty"`
	Parsers  fluentbitconfig.Plugins `yaml:"parsers,omitempty"`
	Multi    fluentbitconfig.Plugins `yaml:"multiline_parsers,omitempty"`
}

// WrappedPipe wraps internal pipeline, adding `multiline_parsers` support.
type WrappedPipe struct {
	fluentbitconfig.Pipeline `yaml:",inline"`
	// Parse additional multiline parsers if they're in this section.
	Multi fluentbitconfig.Plugins `yaml:"multiline_parsers,omitempty"`
}

func dedupeParsers(parsers fluentbitconfig.Plugins) fluentbitconfig.Plugins {
	out := make(fluentbitconfig.Plugins, 0, len(parsers))
	for _, parser := range parsers {
		seen := false
		for _, seenParser := range out {
			// NB: inefficient to compare against every other plugin, but there
			// won't be too many of these and it's useful to keep the order.
			if seenParser.Equal(parser) {
				seen = true
				break
			}
		}

		if !seen {
			out = append(out, parser)
		}
	}

	return out
}
