package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

func start_user_auth(data []byte) string {
	response := GrantRequestResponse{}

	response.Event.Header = HeaderStruct{
		NameSpace:      "Alexa.Authorization",
		Name:           "AcceptGrant.Response",
		PayloadVersion: "3",
		MessageId:      uuid.New().String(),
	}

	message := GrantRequest{}
	json.Unmarshal(data, &message)

	// All we care about is the code
	code := message.Directive.Payload.Grant.Code

	// If we failed to get the token
	_, ok := do_user_auth("authorization_code", "code", code)
	if !ok {
		response.Event.Header.Name = "ErrorResponse"
		response.Event.Payload.Type = "ACCEPT_GRANT_FAILED"
		response.Event.Payload.Message = "We failed to get a token"
	}

	json, err := json.Marshal(response)
	if err != nil {
		log.Println("Err: " + err.Error())
		return ""
	}
	return string(json)
}

// Call this before any use of the user's authn token; we'll refresh
// if needed
func refresh_token() string {
	data := get_button_by_id(DB_TOKEN_AUTH)
	if data.Name == "" {
		return ""
	}

	token := AuthResponse{}
	json.Unmarshal([]byte(data.Name), &token)

	// If we're within the life then we don't need to refresh
	// (60 seconds leeway)
	if token.TimeSet+token.ExpiresIn > time.Now().Unix()-60 {
		log.Println("No refresh needed")
		return data.Name
	}

	log.Println("Refreshing auth token")
	newtoken, ok := do_user_auth("refresh_token", "refresh_token", token.RefreshToken)
	if !ok {
		return data.Name
	} else {
		return newtoken
	}
}

func do_user_auth(grant_type, typestr, value string) (string, bool) {

	// We need to convert this into a token.  We need some information
	// from DynamoDB (client ID/secret)
	client_id := get_button_name(DB_TOKEN_ALEXA_ID)
	client_secret := get_button_name(DB_TOKEN_ALEXA_SECRET)

	if client_id == "" || client_secret == "" {
		log.Println("Ensure you have setclientid and setclientsecret before enabling skill")
		return "", false
	}

	log.Println("Attempting to get tokens for code: " + value)
	request := url.Values{
		"grant_type":    {grant_type},
		"client_id":     {client_id},
		"client_secret": {client_secret},
		typestr:         {value},
	}

	resp, err := http.PostForm("https://api.amazon.com/auth/o2/token", request)

	if err != nil {
		log.Println("Error getting code: " + err.Error())
		return "", false
	}

	if resp.StatusCode != http.StatusOK {
		return "", false
	}

	log.Println(resp.StatusCode)

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	log.Println("Body response: " + string(body))

	token := AuthResponse{}
	json.Unmarshal(body, &token)
	if token.AccessToken == "" {
		log.Println("Did not parse response")
		return "", false
	}
	token.TimeSet = time.Now().Unix()

	result, _ := json.Marshal(token)
	set_button_name(DB_TOKEN_AUTH, string(result))
	return string(result), true
}
