package main

import (
	"encoding/base64"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"log"
)

type MyRequest struct {
	Command  string `json:"command"`
	Param1   string `json:"param1"`
	Param2   string `json:"param2"`
}

type MyResponse struct {
	Message string `json:"Answer:"`
}

func api_wrapper(data []byte) string {
	response_str := handle_api(data)

	// Convert response string to right structure
	response := MyResponse{Message: response_str}
	json, _ := json.Marshal(response)

	return string(json)
}

func handle_api(data []byte) string {
	log.Println(string(data))

	// Parse the incoming message
	message := events.APIGatewayProxyRequest{}
	err := json.Unmarshal(data, &message)
	if err != nil {
		log.Println("Could not parse request: " + err.Error())
		return err.Error()
	}

	// We must have an authorizations header and it must match the
	// password
	our_pswd := get_button_name(DB_TOKEN_PSWD)
	this_pswd := message.Headers["authorization"]

	if this_pswd != our_pswd {
		return "Bad Authorization password"
	}

	// This needs to be a POST request.  But I'm not sure the data
	// structure from the lambda-go/event is parsed properly, so we'll
	// just test for an empty message body

	query := message.Body
	log.Println("Body is: " + query)

	if query == "" {
		return "Error: Must be a POST request with JSON body"
	}

	// This is probably always true but we'll do it properly.
	if message.IsBase64Encoded {
		query2, err := base64.StdEncoding.DecodeString(query)
		if err != nil {
			return "Bad base64 encoded body: " + err.Error()
		}
		query = string(query2)
		log.Println("Decoded body: " + query)
	}

	req := MyRequest{}
	err = json.Unmarshal([]byte(query), &req)
	if err != nil {
		return "Bad JSON passed: " + err.Error()
	}

	return command_mode(req.Command, []string{req.Param1, req.Param2})
}
