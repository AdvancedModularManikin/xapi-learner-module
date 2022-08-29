
package ammlib

import (
   "strings"
)

type PhysiologyModification struct {
   ID, EventId, EventType, Data string
}

var physiologyModificationEventDispatcher map[string]func(pm PhysiologyModification) = make(map[string]func(pm PhysiologyModification))

func RegisterPhysiologyModificationEvent (funcName string, f func(pm PhysiologyModification)) (Error) {
   _, ok := physiologyModificationEventDispatcher[funcName]
   if ok {
      return error1()
   } else {
      physiologyModificationEventDispatcher[funcName] = f
      return error0()
   }
}

func UnregisterPhysiologyModificationEvent (funcName string) (Error) {
   _, ok := physiologyModificationEventDispatcher[funcName]
   if ok {
      return error2()
   } else {
      delete(physiologyModificationEventDispatcher, funcName)
      return error0()
   }
}

func invokePhysiologyModificationEventDispatcher (pm PhysiologyModification) {
   for _, v := range physiologyModificationEventDispatcher {
      /// Need error handling incase rogue function crashes the whole dispatcher.
      v(pm)
   }
}

func parsePhysiologyModificationString (s string) (PhysiologyModification) {
   /// [AMM_Physiology_Modification]id=00000000-0000-0000-0000-000000000000;event_id=00000000-0000-0000-0000-000000000000;type=Command line test;location=;participant_id=;payload=Command line test

   substrings := strings.Split(s[29:], ";")

   var phym PhysiologyModification

   for _, param := range substrings {
      if (strings.Contains(param, "event_id")) {
         phym.EventId = param[9:]
      }
   }

   return phym
}
