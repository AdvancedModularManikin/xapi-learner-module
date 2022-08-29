
package ammlib

import (
   // "strings"
)

type ModuleConfiguration struct {
   Name, ModuleId, EducationalEncounter, Timestamp, CapabilitiesConfiguration string
}

var moduleConfigurationEventDispatcher map[string]func(mc ModuleConfiguration) = make(map[string]func(mc ModuleConfiguration))


func RegisterModuleConfigurationEvent (funcName string, f func(mc ModuleConfiguration)) (Error) {
   _, ok := moduleConfigurationEventDispatcher[funcName]
   if ok {
      return error1()
   } else {
      moduleConfigurationEventDispatcher[funcName] = f
      return error0()
   }
}

func UnregisterModuleConfigurationEvent (funcName string) (Error) {
   _, ok := moduleConfigurationEventDispatcher[funcName]
   if ok {
      return error2()
   } else {
      delete(moduleConfigurationEventDispatcher, funcName)
      return error0()
   }
}

func invokeModuleConfigurationEventDispatcher (mc ModuleConfiguration) {
   for _, v := range moduleConfigurationEventDispatcher {
      /// Need error handling incase rogue function crashes the whole dispatcher.
      v(mc)
   }
}

func parseModuleConfigurationString (s string) (ModuleConfiguration) {
   /// [AMM_Render_Modification]id=00000000-0000-0000-0000-000000000000;event_id=00000000-0000-0000-0000-000000000000;type=Command line test;location=;participant_id=;payload=Command line test

   // substrings := strings.Split(s[29:], ";")

   var mcof ModuleConfiguration

   // for _, param := range substrings {
   // }

   return mcof
}
