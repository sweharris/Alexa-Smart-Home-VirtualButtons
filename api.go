package main

import (
	"encoding/base64"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"log"
	"strings"
)

type MyRequest struct {
	Command string `json:"command"`
	Param1  string `json:"param1"`
	Param2  string `json:"param2"`
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
	var auth, cmd, param1, param2 string

	// Parse the incoming message
	message := events.APIGatewayProxyRequest{}
	err := json.Unmarshal(data, &message)
	if err != nil {
		log.Println("Could not parse request: " + err.Error())
		return err.Error()
	}

	// Now the "correct" way of calling this API is via a POST
	// with an Authorization header and with a JSON body.
	//
	// However it _could_ also be called with a GET and with
	// a cmd=auth/command/parm1/parm2 string
	// We'll only allow GET if insecure mode is set
	insecure := get_button_name(DB_TOKEN_INSECURE)

	query := message.QueryStringParameters["cmd"]
	if query != "" && insecure == "insecure" {
		// Let's use the query string
		// This append() will ensure we have 4 parts
		cmdstr := append(strings.Split(query, "/"), "", "", "", "")
		auth = cmdstr[0]
		cmd = cmdstr[1]
		param1 = cmdstr[2]
		param2 = cmdstr[3]
		log.Println("Command=" + cmd + ", param1=" + param1 + ", param2=" + param2)
	} else {
		// POST request (we hope).  Use the header and body
		auth = message.Headers["authorization"]

		query = message.Body
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

		cmd = req.Command
		param1 = req.Param1
		param2 = req.Param2
	}

	// Ensure the password is correct
	our_pswd := get_button_name(DB_TOKEN_PSWD)

	if auth != our_pswd {
		return "Bad Authorization password"
	}

	return command_mode(cmd, []string{param1, param2})
}
