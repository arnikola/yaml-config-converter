package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
)

var testData = `
parsers:
  - name: year_parser
    format: regex
    time_key: _time
    regex: ^
    time_format: '%Y'
multiline_parsers:
  - name: some:log
    type: regex
    flush_timeout: 1000
    rules:
      - state: start_state
        regex: '/([a-zA-Z]+ \d+ \d+\:\d+\:\d+)(.*)/'
        next_state: cont
      - state: cont
        regex: '^java.*/'
        next_state: cont
  - name: given:type
    type: regex
    parser: foo
    key_content: bar
pipeline:
    parsers:
      - name: hour_parser
        format: regex
        time_key: _time
        regex: ^
        time_format: '%H'
    multiline_parsers:
    - name: other:log
      type: regex
      flush_timeout: 3000
      rules:
      - state: start_state
        regex: '/([a-zA-Z]+/'
        next_state: cont
      - state: cont
        regex: '^golang.*/'
        next_state: cont
    inputs:
      - Name: dummy
        Tag: dummy.data
        Dummy: '{"data":"100 0.5 true example", "key1":"value1", "key2":"value2"}'
    filters:
     - Name: parser
       Match: dummy.*
       Key_Name: data
       Parser: dummy_test
       Reserve_Data: "On"
       Preserve_Key: "On"
    outputs:
      - Name: stdout
        Match: '*'
`

var expected = `
[INPUT]
    Name  dummy
    Tag   dummy.data
    Dummy {"data":"100 0.5 true example", "key1":"value1", "key2":"value2"}
[PARSER]
    name        hour_parser
    format      regex
    time_key    _time
    regex       ^
    time_format %H
[PARSER]
    name        year_parser
    format      regex
    time_key    _time
    regex       ^
    time_format %Y
[FILTER]
    Name         parser
    Match        dummy.*
    Key_Name     data
    Parser       dummy_test
    Reserve_Data On
    Preserve_Key On
[OUTPUT]
    Name  stdout
    Match *
[MULTILINE_PARSER]
    name          other:log
    type          regex
    flush_timeout 3000
    rule          "start_state" "/([a-zA-Z]+/" "cont"
    rule          "cont"        "^golang.*/"   "cont"
[MULTILINE_PARSER]
    name          some:log
    type          regex
    flush_timeout 1000
    rule          "start_state" "/([a-zA-Z]+ \d+ \d+\:\d+\:\d+)(.*)/" "cont"
    rule          "cont"        "^java.*/"                            "cont"
[MULTILINE_PARSER]
    name        given:type
    type        regex
    parser      foo
    key_content bar
`

func TestPrintConfig(t *testing.T) {
	actual, err := printConfig(testData)
	require.NoError(t, err)
	require.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(actual))
}

func TestLua(t *testing.T) {
	someLua := `
function some_func(tag, timestamp, record)
  local os_date = os.date("*t", timestamp)
  local offset = 0
  if os_date.isdst then
      offset = 11 * 3600  -- AEDT (UTC+11)
  else
      offset = 10 * 3600  -- AEST (UTC+10)
  end
  record["time"] = os.date("%Y-%m-%d %H:%M:%S", timestamp + offset) .. "Z"
  return 1, timestamp, record
end
`
	min, err := minifyLua(someLua)
	require.NoError(t, err)
	fmt.Println(min)

	beautify, err := beautifyLua(min)
	require.NoError(t, err)
	fmt.Println(beautify)
}

func minifyLua(inLua string) (string, error) {
	return mutateLua(inLua, "minify")
}

func beautifyLua(inLua string) (string, error) {
	return mutateLua(inLua, "beautify")
}

func mutateLua(inLua, verb string) (string, error) {
	l := lua.NewState(
		lua.Options{},
	)
	defer l.Close()

	f, err := l.LoadFile("minify.lua")
	if err != nil {
		return "", err
	}

	l.Push(f)
	args := l.CreateTable(3, 3)
	args.Append(lua.LString(verb))
	args.Append(lua.LString(inLua))
	l.SetGlobal("args", args)

	err = l.PCall(0, lua.MultRet, nil)
	if err != nil {
		return "", err
	}

	return l.GetGlobal("outputStr").String(), nil
}
