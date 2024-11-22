package main

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"

	fluentbitconfig "github.com/calyptia/go-fluentbit-config/v2"
	"github.com/calyptia/go-fluentbit-config/v2/property"
)

/*
// Taken from
// https://github.com/chronosphereio/calyptia-go-fluentbit-config/blob/main/classic.go#L128
*/

func writePlugins(
	sb io.Writer, kind string, plugins fluentbitconfig.Plugins,
) error {
	for _, plugin := range plugins {
		if err := writeProps(sb, kind, plugin.Properties); err != nil {
			return err
		}
	}

	return nil
}

func writeProps(
	sb io.Writer, kind string, props property.Properties,
) error {
	if len(props) == 0 {
		return nil
	}

	_, err := fmt.Fprintf(sb, "[%s]\n", kind)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(sb, 0, 4, 1, ' ', 0)
	for _, p := range props {
		if s, ok := p.Value.([]any); ok {
			for _, v := range s {
				converted := strings.TrimSuffix(stringFromAny(v), "\n")
				_, err := fmt.Fprintf(tw, "    %s\t%s\n", p.Key, converted)
				if err != nil {
					return err
				}
			}
		} else {
			_, err := fmt.Fprintf(tw, "    %s\t%s\n", p.Key, stringFromAny(p.Value))
			if err != nil {
				return err
			}
		}
	}

	return tw.Flush()
}

// isFloatInt reports whether a float is an integer number
// with no fractional part.
func isFloatInt[F float32 | float64](f F) bool {
	switch t := any(f).(type) {
	case float32:
		return t == float32(int32(f))
	case float64:
		return t == float64(int64(f))
	}
	return false
}

func fmtFloat[F float32 | float64](f F) string {
	s := fmt.Sprintf("%f", f)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}

func int32FromAny(v any) (int32, bool) {
	if v == nil {
		return 0, false
	}

	switch v := v.(type) {
	case int:
		if int(int32(v)) == v {
			return int32(v), true
		}
	case int8:
		if int8(int32(v)) == v {
			return int32(v), true
		}
	case int16:
		if int16(int32(v)) == v {
			return int32(v), true
		}
	case int32:
		return v, true
	case int64:
		if int64(int32(v)) == v {
			return int32(v), true
		}
	case uint:
		if uint(int32(v)) == v {
			return int32(v), true
		}
	case uint16:
		if uint16(int32(v)) == v {
			return int32(v), true
		}
	case uint32:
		if uint32(int32(v)) == v {
			return int32(v), true
		}
	case uint64:
		if uint64(int32(v)) == v {
			return int32(v), true
		}
	case float32:
		if float32(int32(v)) == v {
			return int32(v), true
		}
	case float64:
		if float64(int32(v)) == v {
			return int32(v), true
		}
	case string:
		if i, err := strconv.ParseInt(v, 10, 32); err == nil {
			return int32(i), true
		}
	}
	return 0, false
}

// stringFromAny -
// TODO: Handle more data types.
func stringFromAny(v any) string {
	switch t := v.(type) {
	case encoding.TextMarshaler:
		if b, err := t.MarshalText(); err == nil {
			return stringFromAny(string(b))
		}
	case fmt.Stringer:
		return stringFromAny(t.String())
	case json.Marshaler:
		if b, err := t.MarshalJSON(); err == nil {
			return stringFromAny(string(b))
		}
	case map[string]any, []any:
		var buff bytes.Buffer
		enc := json.NewEncoder(&buff)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(t); err == nil {
			return buff.String()
		}
	case float32:
		if isFloatInt(t) {
			return strconv.FormatInt(int64(t), 10)
		}
		return fmtFloat(t)
	case float64:
		if isFloatInt(t) {
			return strconv.FormatInt(int64(t), 10)
		}
		return fmtFloat(t)
	case string:
		if strings.Contains(t, "\n") {
			return fmt.Sprintf("%q", t)
		}

		if t == "" {
			return `""`
		}

		return t
	}

	return stringFromAny(fmt.Sprintf("%v", v))
}
