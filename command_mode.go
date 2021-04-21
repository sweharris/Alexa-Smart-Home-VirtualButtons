package main

// This file supports the "-command" functions
// There's not really any sanity checks here.  I'm just gonna let
// the GO runtime panic() if the user doesn't provide enough parameters

import (
	"fmt"
	"os"
	"sort"
	"strconv"
)

func dump_buttons(buttons []Button) string {
	sort.Slice(buttons, func(i, j int) bool {
		return buttons[i].ButtonID < buttons[j].ButtonID
	})

	var res string
	var state string
	for _, b := range buttons {
		state = "OFF"
		if b.State == 1 {
			state = "ON "
		}
		res += fmt.Sprintf("%6d %s %s\n", b.ButtonID, state, b.Name)
	}
	return res
}

func command_mode(cmd string, args []string) string {
	switch cmd {
	// List the buttons defined
	case "getbuttons":
		return dump_buttons(get_buttons(false))

	// Define buttons
	case "setname":
		i, _ := strconv.Atoi(args[0])
		set_button_name(i, args[1])
		return get_button_name(i)
	case "setstate":
		i, _ := strconv.Atoi(args[0])
		j, _ := strconv.Atoi(args[1])
		set_button_state(i, j)
		return get_button_name(i)
	case "setstatebyname":
		button := get_button_by_name(args[0])
		i := button.ButtonID
		j, _ := strconv.Atoi(args[1])
		return set_button_state(i, j)
	case "togglestate":
		i, _ := strconv.Atoi(args[0])
		toggle_button_state(i)
		return get_button_name(i)
	case "togglestatebyname":
		button := get_button_by_name(args[0])
		i := button.ButtonID
		toggle_button_state(i)
		return get_button_name(i)
	case "pushcontact":
		i, _ := strconv.Atoi(args[0])
		push_update(i, 1)
		return set_button_state(i, 0)
	case "pushcontactbyname":
		// This will push an update to the API to make Alexa
		// think the contact is open, then set the DynamoDB
		// value to 0 (closed), which itself pushes another update
		// This can be used to trigger a routine on "open".
		button := get_button_by_name(args[0])
		i := button.ButtonID
		push_update(i, 1)
		return set_button_state(i, 0)
	case "deletebutton":
		i, _ := strconv.Atoi(args[0])
		delete_button(i)
		return get_button_name(i)

	// Manage the web password
	case "setpasswd":
		set_button_name(DB_TOKEN_PSWD, args[0])
		return get_button_name(DB_TOKEN_PSWD)
	case "getpasswd":
		return get_button_name(DB_TOKEN_PSWD)

	// Initial setup for the Auth privileges
	case "setclientid":
		set_button_name(DB_TOKEN_ALEXA_ID, args[0])
		return get_button_name(DB_TOKEN_ALEXA_ID)
	case "setclientsecret":
		set_button_name(DB_TOKEN_ALEXA_SECRET, args[0])
		return get_button_name(DB_TOKEN_ALEXA_SECRET)

	// Verify they're set correctly
	case "getclientid":
		return get_button_name(DB_TOKEN_ALEXA_ID)
	case "getclientsecret":
		return get_button_name(DB_TOKEN_ALEXA_SECRET)

	// Debugging stuff you probably don't want to call
	case "getbuttonsall":
		return dump_buttons(get_buttons(true))
	case "gettoken":
		return get_button_name(DB_TOKEN_AUTH)
	case "settoken":
		set_button_name(DB_TOKEN_AUTH, args[0])
		return get_button_name(DB_TOKEN_AUTH)
	case "refreshtoken":
		refresh_token()
		return get_button_name(DB_TOKEN_AUTH)
	case "discovery":
		return discovery_response()
	case "statereport":
		return state_response("xxx", args[0])
	case "pushupdate":
		i, _ := strconv.Atoi(args[0])
		j, _ := strconv.Atoi(args[1])
		return push_update(i, j)
	case "lambda":
		data := os.Getenv("TEST")
		if data == "" {
			fmt.Println("Set TEST variable")
			os.Exit(255)
		}

		rest := process_lambda([]byte(data))
		return string(rest)
	default:
		return "Unknown command: " + cmd
	}
	return ""
}
