package main

// This module will talk to DynamoDB and get a list of all the defined
// buttons and return them in an array structure.
//
// MAGIC!  Button 100000+ contains state information
// Don't define a button with these IDs!
//
// Alexa limits you to maybe 100 (or 300?) buttons

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"log"
	"strconv"
)

// There must be a better way of doing this!
const (
	DB_MAX_VAL            = 99999
	DB_TOKEN_PSWD         = 100000
	DB_TOKEN_AUTH         = 100001
	DB_TOKEN_ALEXA_ID     = 100002
	DB_TOKEN_ALEXA_SECRET = 100003

	TABLE_NAME = "Smart_Home_Virtual_Buttons"
)

// This is all we care about for a button; the ID and the name
type Button struct {
	ButtonID int    `json:"buttonid"`
	Name     string `json:"buttonname"`
	State    int    `json:"buttonstate"`
}

// This will query DynamoDB and get a list of all the buttons.
func get_buttons(all bool) []Button {

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	result, err := svc.Scan(&dynamodb.ScanInput{
		TableName: aws.String(TABLE_NAME),
	})

	if err != nil {
		log.Fatalf("Got error calling Scan: %s", err)
		return nil
	}

	// For each button add them to the slice
	var buttons []Button
	for c := 0; c < len(result.Items); c++ {
		b := Button{}
		dynamodbattribute.UnmarshalMap(result.Items[c], &b)
		// If the button has no name then set a default one
		if b.Name == "" {
			b.Name=fmt.Sprintf("switch-%06d", b.ButtonID)
		}
		if all || b.ButtonID <= DB_MAX_VAL {
			buttons = append(buttons, b)
		}
	}

	return buttons
}

func get_button_by_id(id int) Button {
	var button Button

	buttons := get_buttons(true)
	for _, v := range buttons {
		if v.ButtonID == id {
			button = v
		}
	}
	return button
}

func get_button_by_name(name string) Button {
	var button Button
	button.ButtonID = -1

	buttons := get_buttons(true)
	for _, v := range buttons {
		if v.Name == name {
			button = v
		}
	}
	return button
}

func get_button_name(id int) string {
	return get_button_by_id(id).Name
}

func set_button_name(id int, token string) string {
	if id < 0 {
		return "Bad ID number: " + strconv.Itoa(id)
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(TABLE_NAME),
		// Update command
		UpdateExpression: aws.String("set buttonname = :t"),
		// Where clause
		Key: map[string]*dynamodb.AttributeValue{
			"buttonid": {N: aws.String(strconv.Itoa(id))},
		},
		// Bind variable
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":t": {S: aws.String(token)},
		},
		ReturnValues: aws.String("UPDATED_NEW"),
	}

	_, err := svc.UpdateItem(input)
	if err != nil {
		return "Error updating button: " + err.Error()
	}

	return "Update Success"
}

func set_button_state(id int, state int) string {
	if id < 0 || id > DB_MAX_VAL {
		return "Bad ID number: " + strconv.Itoa(id)
	}

	// Push an immediate update to Alexa
	push_update(id, state)

	// Update DynamoDB
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(TABLE_NAME),
		// Update command
		UpdateExpression: aws.String("set buttonstate = :t"),
		// Where clause
		Key: map[string]*dynamodb.AttributeValue{
			"buttonid": {N: aws.String(strconv.Itoa(id))},
		},
		// Bind variable
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":t": {N: aws.String(strconv.Itoa(state))},
		},
		ReturnValues: aws.String("UPDATED_NEW"),
	}

	_, err := svc.UpdateItem(input)
	if err != nil {
		return "Error updating button: " + err.Error()
	}

	return "Update Success"
}

func toggle_button_state(id int) string {
	button := get_button_by_id(id)
	newstate := 1
	if button.State != 0 {
		newstate = 0
	}
	return set_button_state(id, newstate)
}

func delete_button(id int) string {

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(TABLE_NAME),
		Key: map[string]*dynamodb.AttributeValue{
			"buttonid": {N: aws.String(strconv.Itoa(id))},
		},
	}

	_, err := svc.DeleteItem(input)
	if err != nil {
		return "Error Deleting button: " + err.Error()
	}

	return "Delete Success"
}
