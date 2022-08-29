package ammlib

import (
	"strings"
)

type EventRecord struct {
	ID, EducationalEncounter, EventAgentType, ParticipantId, EventType, Data string
	Timestamp                                                                uint64
	FmaLocation                                                              interface{}
}

var eventRecordEventDispatcher map[string]func(er EventRecord) = make(map[string]func(er EventRecord))

func RegisterEventRecordEvent(funcName string, f func(er EventRecord)) Error {
	_, ok := eventRecordEventDispatcher[funcName]
	if ok {
		return error1()
	} else {
		eventRecordEventDispatcher[funcName] = f
		return error0()
	}
}

func UnregisterEventRecordEvent(funcName string) Error {
	_, ok := eventRecordEventDispatcher[funcName]
	if ok {
		return error2()
	} else {
		delete(eventRecordEventDispatcher, funcName)
		return error0()
	}
}

func invokeEventRecordEventDispatcher(er EventRecord) {
	for _, v := range eventRecordEventDispatcher {
		/// Need error handling incase rogue function crashes the whole dispatcher.
		v(er)
	}
}

func parseEventRecordString(s string) EventRecord {
	// [AMM_EventRecord]id=00000000-0000-0000-0000-000000000000;type=Command line test;location=Command line test;participant_id=;participant_type=Leaner;data=Command line test;

	substrings := strings.Split(s[17:], ";")

	var evtr EventRecord

	for _, param := range substrings {
		if strings.Contains(param, "participant_id=") {
			evtr.ParticipantId = param[15:]
			continue
		}
		if strings.Contains(param, "mid=") {
			continue
		}
		if strings.Contains(param, "id=") {
			evtr.ID = param[3:]
			continue
		}
		if strings.Contains(param, "participant_type=") {
			continue
		}
		if strings.Contains(param, "type=") {
			evtr.EventType = param[5:]
		}
	}

	return evtr
}
