package main

import lua "github.com/yuin/gopher-lua"

func minifyLua(inLua string) (string, error) {
	return mutateLua(inLua, "minify")
}

func beautifyLua(inLua string) (string, error) {
	return mutateLua(inLua, "beautify")
}

func mutateLua(inLua, verb string) (string, error) {
	l := lua.NewState()
	defer l.Close()

	// Load minifier
	f, err := l.LoadFile("minify.lua")
	if err != nil {
		return "", err
	}

	l.Push(f)

	// Inject args
	args := l.CreateTable(2, 2)
	args.Append(lua.LString(verb))
	args.Append(lua.LString(inLua))
	l.SetGlobal("args", args)

	// Run minifier
	err = l.PCall(0, lua.MultRet, nil)
	if err != nil {
		return "", err
	}

	// Retrieve output
	return l.GetGlobal("outputStr").String(), nil
}
