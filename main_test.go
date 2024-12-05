package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
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
		offset = 11 * 3600
	else
		offset = 10 * 3600
	end
	record["time"] = os.date("%Y-%m-%d %H:%M:%S", timestamp + offset).."Z"
	return 1, timestamp, record
end
`
	minEx := `function some_func(tag,timestamp,record)local os_date=os.date("*t",timestamp)local offset=0 if os_date.isdst then offset=11*3600 else offset=10*3600 end record["time"]=os.date("%Y-%m-%d %H:%M:%S",timestamp+offset).."Z"return 1,timestamp,record end`
	min, err := minifyLua(someLua)
	require.NoError(t, err)
	require.Equal(t, strings.TrimSpace(minEx), strings.TrimSpace(min))

	beautify, err := beautifyLua(min)
	require.NoError(t, err)
	require.Equal(t, strings.TrimSpace(someLua), strings.TrimSpace(beautify))
}

func TestLuaParser(t *testing.T) {
	data := `
pipeline:
  filters:
    - name: lua
      match: something
      code: |
        function some_func(tag, timestamp, record)
            local os_date = os.date("*t", timestamp)
            local offset = 0
            if os_date.isdst then
                offset = 11 * 3600
            else
                offset = 10 * 3600
            end
            record["time"] = os.date("%Y-%m-%d %H:%M:%S", timestamp + offset) .. "Z"
            return 1, timestamp, record
        end
      call: some_func
      `

	expected := `
  [FILTER]
    name  lua
    match something
    code  function some_func(tag,timestamp,record)local os_date=os.date("*t",timestamp)local offset=0 if os_date.isdst then offset=11*3600 else offset=10*3600 end record["time"]=os.date("%Y-%m-%d %H:%M:%S",timestamp+offset).."Z"return 1,timestamp,record end
    call  some_func
  `

	actual, err := printConfig(data)
	require.NoError(t, err)
	require.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(actual))
}
