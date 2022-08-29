
package ammlib

import (
   // "fmt"
   "strings"
   "encoding/xml"
)

type RenderModification struct {
   ID, ParticipantId, EventId, EventType string
   Payload RenderModXMLPayload
}

type RenderModXMLPayload struct {
   // XMLName xml.Name `xml:"RenderModification"`
   RenderType string `xml:"type,attr"`
}

var RenderModificationTypeBlacklist []string = make([]string, 0)

var renderModificationEventDispatcher map[string]func(rm RenderModification) = make(map[string]func(rm RenderModification))


func RegisterRenderModificationEvent (funcName string, f func(rm RenderModification)) (Error) {
   _, ok := renderModificationEventDispatcher[funcName]
   if ok {
      return error1()
   } else {
      renderModificationEventDispatcher[funcName] = f
      return error0()
   }
}

func UnregisterRenderModificationEvent (funcName string) (Error) {
   _, ok := renderModificationEventDispatcher[funcName]
   if ok {
      return error2()
   } else {
      delete(renderModificationEventDispatcher, funcName)
      return error0()
   }
}

func invokeRenderModificationEventDispatcher (rm RenderModification) {
   for _, v := range renderModificationEventDispatcher {
      /// Need error handling incase rogue function crashes the whole dispatcher.
      v(rm)
   }
}

func parseRenderModificationString (s string) (RenderModification, bool) {
   /// [AMM_Render_Modification]id=00000000-0000-0000-0000-000000000000;event_id=00000000-0000-0000-0000-000000000000;type=Command line test;location=;participant_id=;payload=Command line test

   /// Blacklist render mods by type.
   for _, phrase := range RenderModificationTypeBlacklist {
      if strings.Contains(s, phrase) {
         return RenderModification{}, false
      }
   }

   substrings := strings.Split(s[29:], ";")

   var redm RenderModification
   var payload RenderModXMLPayload

   for _, param := range substrings {
      // if strings.Contains(param, "event_id") {
      //    redm.EventId = param[9:]
      // }
      if strings.Contains(param, "participant_id") {
         redm.ParticipantId = param[15:]
      }
      if strings.Contains(param, "payload") {
         // fmt.Println("payload found")
         xmlStr := param[8:]
         // fmt.Println("xml string:", xmlStr)
         xml.Unmarshal([]byte(xmlStr), &payload)
         // fmt.Println("struct data:", payload)
         redm.Payload = payload
      } else {
         // fmt.Println("No payload found")
      }
   }

   // fmt.Println("Render Mod Parse:", redm)

   return redm, true
}
