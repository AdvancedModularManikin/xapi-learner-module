/// Built with Go 1.15.6.

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync/atomic"

	// "strings"
	// "net"
	"net/http"
	// "encoding/base64"
	"bytes"
	"encoding/json"

	// "bufio"
	"io/ioutil"
	"time"
	amm "vcom3d/ammlib"
	// "strconv"
)

type Config struct {
	LrsHost         string `json:"LrsHost"`
	LrsXapiEndpoint string `json:"LrsXapiEndpoint"`
	LrsUser         string `json:"LrsUser"`
	LrsPswd         string `json:"LrsPswd"`

	AmmTcpBridgeAddress       string   `json:"AmmTcpBridgeAddress"`
	AmmRenderModTypeBlackList []string `json:"AmmRenderModTypeBlackList"`
}

// var asmtMap map[string]amm.Assessment             = make(map[string]amm.Assessment)
var evtrMap map[string]amm.EventRecord = make(map[string]amm.EventRecord)
var oevtrMap map[string]amm.OmittedEvent = make(map[string]amm.OmittedEvent)

// var phymMap map[string]amm.PhysiologyModification = make(map[string]amm.PhysiologyModification)
// var redmMap map[string]amm.RenderModification     = make(map[string]amm.RenderModification)

var config Config
var rendermodXapiParings map[string]interface{}

var exPath string
var debugCounter uint64

func main() {
	fmt.Println("===== xAPI AMM Module =====\n")

	ex, err := os.Executable()
	check(err)
	exPath := filepath.Dir(ex)
	fmt.Println(exPath)

	qPath := filepath.Join(exPath, "xapi-queue")
	err = os.MkdirAll(qPath, os.ModePerm)
	if err != nil {
		log.Println(err)
	}

	configData, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal("Error reading config file: ", err)
	}

	{
		err := json.Unmarshal(configData, &config)
		if err != nil {
			log.Fatal("Error unmarshalling config json: ", err)
		}
	}

	rendermodXapiParingsData, err := ioutil.ReadFile("xapi-mappings.json")
	if err != nil {
		log.Fatal("Error reading parings file: ", err)
	}

	{
		err := json.Unmarshal(rendermodXapiParingsData, &rendermodXapiParings)
		if err != nil {
			log.Fatal("Error unmarshalling parings json: ", err)
		}
	}

	fmt.Printf("Configuration: %+v\n\n", config)
	fmt.Printf("xAPI Render Mod Parings: %+v\n\n", rendermodXapiParings)

	amm.RegisterAssessmentEvent("onAssessment", onAssessment)
	amm.RegisterEventRecordEvent("onEventRecord", onEventRecord)
	amm.RegisterOmittedEventEvent("onOmittedEvent", onOmittedEvent)
	amm.RegisterPhysiologyModificationEvent("onPhysiologyModification", onPhysiologyModification)
	amm.RegisterRenderModificationEvent("onRenderModification", onRenderModification)
	amm.RegisterModuleConfigurationEvent("onModuleConfiguration", onModuleConfiguration)
	amm.RegisterCommandEvent("onCommand", onCommand)
	amm.RegisterDisconnectEvent("onDisconnect", onDisconnect)

	ammConfig := amm.UserModuleConfig{
		Name:          "XApiAmmModule",
		Manufacturer:  "Vcom3D",
		Model:         "XApiAmmModule",
		SerialNumber:  "1f7b79424be713a34097f5dba4e189b1c14d49b9",
		ModuleVersion: "1.1.3",
	}

	/// Filter out noisy render mod types
	for _, v := range config.AmmRenderModTypeBlackList {
		amm.RenderModificationTypeBlacklist = append(amm.RenderModificationTypeBlacklist, v)
	}

	/// Registering for AMM topics instructs the library to create a subscriber of that type.
	/// Be sure to call register for all desired events before calling Connect.
	amm.ServerAddress = config.AmmTcpBridgeAddress
	amm.Connect(ammConfig)

	var userInput string
	for {
		time.Sleep(500 * time.Millisecond)

		_, err := fmt.Scan(&userInput)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func nextDebugId() string {
	return fmt.Sprintf("%d", atomic.AddUint64(&debugCounter, 1))
}

func buildFileName() string {
	return time.Now().Format("20060102150405") + "_" + nextDebugId()
}

func onDisconnect() {
	log.Println("Disconnected from AMM TCP Bridge")
}

func onAssessment(asmt amm.Assessment) {
	log.Println("Assessment Received", asmt)
	t, _ := time.Now().UTC().MarshalText()

	if tEv, ok := evtrMap[asmt.EventId]; ok {
		log.Println("Found matching event record")
		verb, object := checkRendermodXapiParings(tEv.EventType)
		xapi := generateXapi(tEv.ParticipantId, object, verb, asmt.Value, string(t))
		queueXapi(xapi)
	} else if tEv, ok := oevtrMap[asmt.EventId]; ok {
		log.Println("Found matching omitted event record")
		verb, object := checkRendermodXapiParings(tEv.EventType)
		xapi := generateXapi(tEv.ParticipantId, object, verb, asmt.Value, string(t))
		queueXapi(xapi)
	} else {
		log.Println("No matching event record, no xAPI statement generated")
		return
	}

}

func onOmittedEvent(oevtr amm.OmittedEvent) {
	//log.Println("Omitted Event Record Received!", oevtr)
	oevtrMap[oevtr.ID] = oevtr
	log.Println("Stored omitted event record using ID ", oevtr.ID)
}

func onEventRecord(evtr amm.EventRecord) {
	//log.Println("Event Record Received!", evtr)
	evtrMap[evtr.ID] = evtr
	log.Println("Stored event record using ID ", evtr.ID)
}

func onPhysiologyModification(pm amm.PhysiologyModification) {
	// log.Println("Physiology Modification Received!\n", pm)
}

func onRenderModification(rm amm.RenderModification) {
	// log.Println("Render Modification Received!\n", rm)
}

func queueXapi(xapi string) {
	fullpath := filepath.Join(exPath, "xapi-queue", buildFileName())

	fd, err := os.Create(fullpath)
	if err != nil {
		log.Print(err)
		return
	}
	//log.Println("Storing xAPI statement in queue", xapi)
	fd.WriteString(xapi)
	defer fd.Close()
}

func processXapiFile(filename string) {

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	// Convert []byte to string and print to screen
	xapi := string(content)
	resp := postXapi(xapi, config.LrsHost, config.LrsHost+config.LrsXapiEndpoint, config.LrsUser, config.LrsPswd)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode == 200 {
		log.Println("Successfully posted xAPI statement, deleting queue file " + filename)
		e := os.Remove(filename)
		if e != nil {
			log.Fatal(e)
		}
	} else {
		log.Println("ERROR: Did not successfully post xAPI statement to LRS, keeping " + filename + " in queue")
	}
}

func processXapiQueue() {
	items, _ := ioutil.ReadDir(filepath.Join(exPath, "xapi-queue"))
	for _, item := range items {
		if item.IsDir() {
			subitems, _ := ioutil.ReadDir(item.Name())
			for _, subitem := range subitems {
				if !subitem.IsDir() {
					filename := filepath.Join(exPath, "xapi-queue", item.Name(), subitem.Name())
					fmt.Println("Processing " + filename)
					processXapiFile(filename)

				}
			}
		} else {
			// handle file there
			filename := filepath.Join(exPath, "xapi-queue", item.Name())
			fmt.Println("Processing " + filename)
			processXapiFile(filename)

		}
	}
}

func onCommand(cm amm.Command) {
	if cm.Message == "ACT=[SYS]PUBLISH_ASSESSMENT" {
		processXapiQueue()
	}
}

func onModuleConfiguration(mc amm.ModuleConfiguration) {
	//log.Println("Module Configuration Received!\n", mc)

}

func checkRendermodXapiParings(rendermodType string) (verb, object string) {
	/// rendermodXapiParings["CHECK_MOUTH"].(map[string]interface{})["parings"].([]interface{})[0].(map[string]interface{})["xAPIObject"]

	val, ok := rendermodXapiParings[rendermodType]
	if !ok {
		return "performed", rendermodType
	}

	rtnVal := val.(map[string]interface{})["parings"].([]interface{})[0].(map[string]interface{})

	return rtnVal["xAPIVerb"].(string), rtnVal["xAPIObject"].(string)
}

func generateXapi(actorName, objectName, verbName, value, timestamp string) string {

	bSuccess := "false"

	if actorName == "" {
		actorName = "System"
	}

	if objectName == "" {
		objectName = "Unknown Object"
	}

	if verbName == "" {
		verbName = "performed"
	}

	if value == "Success" {
		bSuccess = "true"
	} else {
		bSuccess = "false"
	}

	xApiStmt := `{
      "actor": {
         "name": "` + actorName + `",
         "mbox": "mailto:email@domain.net",
         "objectType": "Agent"
      },
      "verb": {
         "id": "http://adlnet.gov/expapi/verbs/performed",
         "display": {
            "en-us": "` + verbName + `"
         }
      },
      "object": {
         "id": "http://domain.net/objects/placeholder",
         "definition": {
            "name": {
               "en-us": "` + objectName + `"
            }
         }
      },
	  "result": {
		 "completion" : true,
		 "response"	  : "` + value + `",
		 "success"    : ` + bSuccess + `
	  },
	  "timestamp": "` + timestamp + `"
   }`

	return xApiStmt
}

func postXapi(xApiStmt, host, url, user, pswd string) *http.Response {
	fmt.Println(xApiStmt)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(xApiStmt)))
	if err != nil {
		log.Println("Error reading request. ", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Host", host)
	req.Header.Set("X-Experience-API-Version", "1.0.3")
	req.SetBasicAuth(user, pswd)

	client := &http.Client{Timeout: time.Second * 10}

	// fmt.Println(req.Header)

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error reading response. ", err)
		return resp
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading body. ", err)
		return resp
	}

	fmt.Printf("%s\n", body)

	return resp
}
