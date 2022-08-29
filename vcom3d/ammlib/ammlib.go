package ammlib

import (
	"fmt"
	"log"
	"net"
	"strings"

	// "net/http"
	"encoding/base64"
	// "encoding/xml"
	// "bytes"
	"bufio"
	// "io/ioutil"
	// "time"
)

type UserModuleConfig struct {
	Name, Manufacturer, Model, SerialNumber, ModuleVersion string
}

var ServerAddress string = "127.0.0.1:9015"

/// k = name, v = minimum version
var Versions map[string]string = make(map[string]string)

func generateCapabilityString(umc UserModuleConfig) string {
	sb := strings.Builder{}
	sb.WriteString(`<AMMModuleConfiguration><module name="`)
	sb.WriteString(umc.Name)
	sb.WriteString(`" manufacturer="`)
	sb.WriteString(umc.Manufacturer)
	sb.WriteString(`" model="`)
	sb.WriteString(umc.Model)
	sb.WriteString(`" serial_number="`)
	sb.WriteString(umc.SerialNumber)
	sb.WriteString(`" module_version="`)
	sb.WriteString(umc.ModuleVersion)
	sb.WriteString(`"><versions>`)
	for k, v := range Versions {
		sb.WriteString(`<data name="`)
		sb.WriteString(k)
		sb.WriteString(`" `)
		sb.WriteString(`minimumVersion="`)
		sb.WriteString(v)
		sb.WriteString(`" />`)
	}
	sb.WriteString(`</versions><capabilities><capability name="`)
	sb.WriteString(umc.Name)
	sb.WriteString(`"><subscribed_topics>`)
	sb.WriteString(`<topic name="AMM_ModuleConfiguration" />`)
	if len(assessmentEventDispatcher) > 0 {
		sb.WriteString(`<topic name="AMM_Assessment" />`)
	}
	if len(eventRecordEventDispatcher) > 0 {
		sb.WriteString(`<topic name="AMM_EventRecord" />`)
	}
	if len(omittedEventEventDispatcher) > 0 {
		sb.WriteString(`<topic name="AMM_OmittedEvent" />`)
	}
	if len(physiologyModificationEventDispatcher) > 0 {
		sb.WriteString(`<topic name="AMM_Physiology_Modification" />`)
	}
	if len(renderModificationEventDispatcher) > 0 {
		sb.WriteString(`<topic name="AMM_Render_Modification" />`)
	}
	if len(moduleConfigurationEventDispatcher) > 0 {
		sb.WriteString(`<topic name="AMM_Module_Configuration" />`)
	}
	if len(CommandEventDispatcher) > 0 {
		sb.WriteString(`<topic name="AMM_Command" />`)
	}
	sb.WriteString(`</subscribed_topics></capability></capabilities></module></AMMModuleConfiguration>`)
	fmt.Println("AMM Module Configuration:", sb.String(), "\n")
	return sb.String()
}

/// ### Disconnect Event

var disconnectEventDispatcher map[string]func() = make(map[string]func())

func RegisterDisconnectEvent(funcName string, f func()) Error {
	_, ok := disconnectEventDispatcher[funcName]
	if ok {
		return error1()
	} else {
		disconnectEventDispatcher[funcName] = f
		return error0()
	}
}

func UnregisterDisconnectEvent(funcName string) Error {
	_, ok := disconnectEventDispatcher[funcName]
	if ok {
		return error2()
	} else {
		delete(disconnectEventDispatcher, funcName)
		return error0()
	}
}

func invokeDisconnectEventDispatcher() {
	for _, v := range disconnectEventDispatcher {
		/// Need error handling incase rogue function crashes the whole dispatcher.
		v()
	}
}

/// ###

var conn net.Conn

func Connect(umc UserModuleConfig) {
	connect([]byte(generateCapabilityString(umc)))
}

func connect(capability []byte) {
	log.Println("Connecting to ", ServerAddress)
	conn, err := net.Dial("tcp", ServerAddress)
	if err != nil {
		log.Fatal(err)
	}
	//defer conn.Close()

	go func(conn *net.Conn) {
		for {
			// fmt.Println("Listening for server...")
			data, _ := bufio.NewReader(*conn).ReadBytes('\n')
			//log.Printf("Received %d bytes\n", len(data))
			parseServerMsg(data)
			if len(data) == 0 {
				invokeDisconnectEventDispatcher()
				(*conn).Close()
				break
			}
		}
	}(&conn)

	fmt.Println("Connected to AMM TCP Bridge.")

	capabilityBase64 := base64.StdEncoding.EncodeToString(capability)
	conn.Write([]byte("CAPABILITY=" + capabilityBase64 + "\n"))

}

func Disconnect() {
	if conn != nil {
		conn.Close()
		invokeDisconnectEventDispatcher()
	}
}

func parseServerMsg(data []byte) {
	// fmt.Println("Server says: " + string(data))
	if strings.Contains(string(data), "AMM_Assessment") {
		asmt := parseAssessmentString(string(data))
		invokeAssessmentEventDispatcher(asmt)

	} else if strings.Contains(string(data), "AMM_EventRecord") {
		evtr := parseEventRecordString(string(data))
		invokeEventRecordEventDispatcher(evtr)
	} else if strings.Contains(string(data), "AMM_OmittedEvent") {
		oevtr := parseOmittedEventString(string(data))
		invokeOmittedEventEventDispatcher(oevtr)
	} else if strings.Contains(string(data), "AMM_Physiology_Modification") {
		phym := parsePhysiologyModificationString(string(data))
		invokePhysiologyModificationEventDispatcher(phym)

	} else if strings.Contains(string(data), "AMM_Render_Modification") {
		redm, ok := parseRenderModificationString(string(data))
		if ok {
			invokeRenderModificationEventDispatcher(redm)
		}

	} else if strings.Contains(string(data), "AMM_Module_Configuration") {
		mcof := parseModuleConfigurationString(string(data))
		invokeModuleConfigurationEventDispatcher(mcof)
	} else if strings.Contains(string(data), "AMM_Command") {
		mcof := parseCommandString(string(data))
		invokeCommandEventDispatcher(mcof)
	} else if strings.Contains(string(data), "ACT=[SYS]") {
		mcof := parseCommandString(string(data))
		invokeCommandEventDispatcher(mcof)
	}

}
