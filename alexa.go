package main

// This module will handle the alexa messages

// I have no idea if this is idiomatic golang or if I'm doing silly
// stuff.  All these sub-structures seem awkward.  But... I think it's
// readable!

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// This builds the DiscoveryResponse JSON.  It's based purely on the
// data in the database.
func discovery_response() string {
	buttons := get_buttons(false)

	response := DiscoveryResponse{}

	// The Header is easy enough
	response.Event.Header = HeaderStruct{
		NameSpace:      "Alexa.Discovery",
		Name:           "Discover.Response",
		PayloadVersion: "3",
		MessageId:      uuid.New().String(),
	}

	// Now we need to build an array of endpoints.
	var endpoints []EndpointResponse

	for _, button := range buttons {
		endpoint := EndpointResponse{
			Description:       button.Name,
			EndpointID:        fmt.Sprintf("switch-%06d", button.ButtonID),
			FriendlyName:      button.Name,
			ManufacturerName:  "Virtual Switch",
			DisplayCategories: []string{"CONTACT_SENSOR"},
		}

		// This is where the magic happens.  We have 3 capabilities
		// The first is just defining an Alexa Interface
		// The second defines the button
		// The third defines "health"
		var cap []CapabilitiesResponse

		cap = append(cap, CapabilitiesResponse{
			Type:      "AlexaInterface",
			Interface: "Alexa",
			Version:   "3",
		})

		// This is the main one.  I'm not sure if there's an
		// easier way of doing this, initializing an array of
		// structures?  Hmm. I'm gonna do it the naive way!

		sup := SupportedResponse{
			Name: "detectionState",
		}
		prop := PropertiesResponse{
			Supported:   []SupportedResponse{sup},
			Proactive:   true,
			Retrievable: true,
		}
		cap = append(cap, CapabilitiesResponse{
			Type:       "AlexaInterface",
			Interface:  "Alexa.ContactSensor",
			Version:    "3",
			Properties: &prop,
		})

		// And now the healthcheck, built out similarly
		// "prop" needs to be a unique variable 'cos it's a pointer
		// in the "Capabilities" structure so changing "prop"
		// would change the earlier entry, so we create "prop2"
		sup = SupportedResponse{
			Name: "connectivity",
		}
		prop2 := PropertiesResponse{
			Supported:   []SupportedResponse{sup},
			Proactive:   true,
			Retrievable: true,
		}
		cap = append(cap, CapabilitiesResponse{
			Type:       "AlexaInterface",
			Interface:  "Alexa.EndpointHealth",
			Version:    "3",
			Properties: &prop2,
		})

		endpoint.Capabilities = cap

		endpoints = append(endpoints, endpoint)
	}
	response.Event.Payload.Endpoints = endpoints

	json, err := json.Marshal(response)
	if err != nil {
		log.Println("Err: " + err.Error())
		return ""
	}
	return string(json)
}

// Asking for the state of the button.  We report on the state of the button
// (NOT_DETECTED or DETECTED)
func state_response(correl string, id string) string {

	// Convert the button id to a button index
	idnumber, _ := strconv.Atoi(strings.TrimPrefix(id, "switch-"))
	button := get_button_by_id(idnumber)
	buttonstate := "DETECTED"
	if button.State == 0 {
		buttonstate = "NOT_" + buttonstate
	}

	response := StateResponse{}

	response.Event.Header = HeaderStruct{
		NameSpace:      "Alexa",
		Name:           "StateReport",
		PayloadVersion: "3",
		MessageId:      uuid.New().String(),
		Correlation:    correl,
	}

	response.Event.Endpoint.EndpointID = id

	// Another array.  But simpler this time.  There's got to be
	// a better way!
	prop := make([]StatePropertiesResponse, 2)

	now := time.Now()
	prop[0] = StatePropertiesResponse{
		NameSpace:    "Alexa.ContactSensor",
		Name:         "detectionState",
		Value:        buttonstate,
		TimeOfSample: now,
		Uncertainty:  0,
	}

	prop[1] = StatePropertiesResponse{
		NameSpace: "Alexa.EndpointHealth",
		Name:      "connectivity",
		Value: struct {
			Value string `json:"value"`
		}{Value: "OK"},
		TimeOfSample: now,
		Uncertainty:  0,
	}

	response.Context.Properties = prop

	json, err := json.Marshal(response)
	if err != nil {
		log.Println("Err: " + err.Error())
		return ""
	}
	return string(json)
}

// Push an update message.

func push_update(id, state int) string {
	log.Println("Pushing a notification to Alexa")

	refresh_token()

	tokenStruct := get_button_name(DB_TOKEN_AUTH)
	token := AuthResponse{}
	json.Unmarshal([]byte(tokenStruct), &token)

	response := ChangeReport{}

	response.Event.Header = HeaderStruct{
		NameSpace:      "Alexa",
		Name:           "ChangeReport",
		PayloadVersion: "3",
		MessageId:      uuid.New().String(),
	}

	response.Event.Endpoint.Scope.Type = "BearerToken"
	response.Event.Endpoint.Scope.Token = token.AccessToken
	response.Event.Endpoint.EndpointID = fmt.Sprintf("switch-%06d", id)

	response.Event.Payload.Change.Cause.Type = "PHYSICAL_INTERACTION"

	prop := make([]StatePropertiesResponse, 1)

	now := time.Now()

	statestr := "DETECTED"
	if state == 0 {
		statestr = "NOT_DETECTED"
	}
	prop[0] = StatePropertiesResponse{
		NameSpace:    "Alexa.ContactSensor",
		Name:         "detectionState",
		Value:        statestr,
		TimeOfSample: now,
		Uncertainty:  0,
	}
	response.Event.Payload.Change.Properties = prop

	prop2 := make([]StatePropertiesResponse, 1)

	prop2[0] = StatePropertiesResponse{
		NameSpace: "Alexa.EndpointHealth",
		Name:      "connectivity",
		Value: struct {
			Value string `json:"value"`
		}{Value: "OK"},
		TimeOfSample: now,
		Uncertainty:  0,
	}
	response.Context.Properties = prop2

	json, _ := json.Marshal(response)
	log.Println(string(json))

	resp, err := http.Post("https://api.amazonalexa.com/v3/events", "application/json", bytes.NewBuffer(json))

	if err != nil {
		log.Println("Error sending update: " + err.Error())
	} else {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)

		log.Println("Body response: " + string(body))
	}
	return "Completed"
}

func handle_alexa(data []byte) string {
	// Parse the incoming message
	message := AlexaMessage{}
	err := json.Unmarshal(data, &message)

	log.Println(string(data))

	if err != nil {
		log.Println("Bad parse: " + err.Error())
		return "Bad JSON: " + err.Error()
	}

	namespace := message.Directive.Header.NameSpace
	name := message.Directive.Header.Name
	correl := message.Directive.Header.Correlation
	id := message.Directive.Endpoint.EndpointID

	log.Println("NameSpace: " + namespace + ", Name: " + name)
	result := ""
	if namespace == "Alexa.Authorization" && name == "AcceptGrant" {
		result = start_user_auth(data)
	} else if namespace == "Alexa.Discovery" && name == "Discover" {
		result = discovery_response()
	} else if namespace == "Alexa" && name == "ReportState" {
		result = state_response(correl, id)
	}

	log.Println("Result of Handler: " + result)
	return result
}
