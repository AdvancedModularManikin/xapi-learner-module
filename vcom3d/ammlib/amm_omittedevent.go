package ammlib

import (
	"strings"
)

type OmittedEvent struct {
	ID, EducationalEncounter, EventAgentType, ParticipantId, EventType, Data string
	Timestamp                                                                uint64
	FmaLocation                                                              interface{}
}

var omittedEventEventDispatcher map[string]func(er OmittedEvent) = make(map[string]func(er OmittedEvent))

func RegisterOmittedEventEvent(funcName string, f func(er OmittedEvent)) Error {
	_, ok := omittedEventEventDispatcher[funcName]
	if ok {
		return error1()
	} else {
		omittedEventEventDispatcher[funcName] = f
		return error0()
	}
}

func UnregisterOmittedEventEvent(funcName string) Error {
	_, ok := omittedEventEventDispatcher[funcName]
	if ok {
		return error2()
	} else {
		delete(omittedEventEventDispatcher, funcName)
		return error0()
	}
}

func invokeOmittedEventEventDispatcher(er OmittedEvent) {
	for _, v := range omittedEventEventDispatcher {
		/// Need error handling incase rogue function crashes the whole dispatcher.
		v(er)
	}
}

func parseOmittedEventString(s string) OmittedEvent {
	// [AMM_OmittedEvent]id=00000000-0000-0000-0000-000000000000;type=Command line test;location=Command line test;participant_id=;participant_type=Leaner;data=Command line test;

	substrings := strings.Split(s[18:], ";")

	var oevtr OmittedEvent

	for _, param := range substrings {
		if strings.Contains(param, "participant_id=") {
			oevtr.ParticipantId = param[15:]
			continue
		}
		if strings.Contains(param, "mid=") {
			continue
		}
		if strings.Contains(param, "id=") {
			oevtr.ID = param[3:]
			continue
		}
		if strings.Contains(param, "participant_type=") {
			continue
		}
		if strings.Contains(param, "type=") {
			oevtr.EventType = param[5:]
		}
	}

	return oevtr
}
