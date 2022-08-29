package ammlib

import (
	"fmt"
	"strings"
)

// "strings"

type Command struct {
	Message string
}

var CommandEventDispatcher map[string]func(mc Command) = make(map[string]func(mc Command))

func RegisterCommandEvent(funcName string, f func(mc Command)) Error {
	_, ok := CommandEventDispatcher[funcName]
	if ok {
		return error1()
	} else {
		CommandEventDispatcher[funcName] = f
		return error0()
	}
}

func UnregisterCommandEvent(funcName string) Error {
	_, ok := CommandEventDispatcher[funcName]
	if ok {
		return error2()
	} else {
		delete(CommandEventDispatcher, funcName)
		return error0()
	}
}

func invokeCommandEventDispatcher(mc Command) {
	for _, v := range CommandEventDispatcher {
		/// Need error handling incase rogue function crashes the whole dispatcher.
		v(mc)
	}
}

func parseCommandString(s string) Command {

	var mcof Command

	parts := strings.Split(s, ";")
	fmt.Println(parts[0])

	mcof.Message = parts[0]

	return mcof
}
