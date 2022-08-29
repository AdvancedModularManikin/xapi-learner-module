package ammlib

import (
	"log"
	"strings"
)

type Assessment struct {
	ID, EventId, Value, Comment string
}

var assessmentEventDispatcher map[string]func(a Assessment) = make(map[string]func(a Assessment))

func RegisterAssessmentEvent(funcName string, f func(a Assessment)) Error {
	_, ok := assessmentEventDispatcher[funcName]
	if ok {
		return error1()
	} else {
		assessmentEventDispatcher[funcName] = f
		return error0()
	}
}

func UnregisterAssementEvent(funcName string) Error {
	_, ok := assessmentEventDispatcher[funcName]
	if ok {
		return error2()
	} else {
		delete(assessmentEventDispatcher, funcName)
		return error0()
	}
}

func invokeAssessmentEventDispatcher(a Assessment) {
	for _, v := range assessmentEventDispatcher {
		/// Need error handling incase rogue function crashes the whole dispatcher.
		v(a)
	}
}

func parseAssessmentString(s string) Assessment {
	// [AMM_Assessment]id=00000000-0000-0000-0000-000000000000;event_id=00000000-0000-0000-0000-000000000000;type=Command line test;location=Command line test;participant_id=;value=Omission Error;comment=Command line test

	substrings := strings.Split(s[16:], ";")

	var asmt Assessment

	for _, param := range substrings {
		if strings.Contains(param, "event_id=") {
			log.Println("Event ID set")
			asmt.EventId = param[9:]
			continue
		}
		if strings.Contains(param, "value=") {
			asmt.Value = param[6:]
		}
	}

	return asmt
}
