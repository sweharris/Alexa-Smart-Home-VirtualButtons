package main

import (
	"context"
	"encoding/json"
	"log"
)

// This is a kludge to try and allow multiple types of trigger

type Handler struct{}

func (h Handler) Invoke(ctx context.Context, data []byte) ([]byte, error) {
	return process_lambda(data), nil
}

func process_lambda(data []byte) []byte {
	log.Println("Handler started")

	// We'll initially just map the incoming data to an array.  This will
	// let us easily look at the fields at the top of the JSON structure
	// and make a guestimate as to the type of call.

	log.Println("Incoming data: ",string(data))

	result := "No results"

	var req map[string]interface{}
	err := json.Unmarshal(data, &req)

	if err != nil {
		result = "Bad parse: " + err.Error()
		log.Println("Bad parse: " + err.Error())
	} else if req["directive"] != nil {
		log.Println("Alexa call")
		result = handle_alexa(data)
	} else if req["routeKey"] != nil || req["httpMethod"] != nil {
		log.Println("API call")
		result = api_wrapper(data)
	}

	log.Println("Handler ended: " + result)

	return []byte(result)
}
