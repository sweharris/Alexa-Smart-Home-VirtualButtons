package main

import "time"

// This defines all the structures used to communicate to/from the Alexa
// trigger and associated endpoints.

// I have no idea if this is idiomatic golang or if I'm doing silly
// stuff.  All these sub-structures seem awkward.  But... I think it's
// readable!

// Common for incoming message and outgoing responses

type HeaderStruct struct {
	NameSpace      string `json:"namespace"`
	Name           string `json:"name"`
	MessageId      string `json:"messageId"`
	PayloadVersion string `json:"payloadVersion"`
	Correlation    string `json:"correlationToken,omitempty"`
}

// The incoming message.  Discovery and State Report messages are
// slightly different, but this should handle both.

type AlexaMessage struct {
	Directive struct {
		Header   HeaderStruct `json:"header"`
		Endpoint struct {
			Scope struct {
				Type  string `json:"type"`
				Token string `json:"token"`
			} `json:"scope"`
			EndpointID string `json:"endpointId"`
		} `json:"endpoint"`
		Payload struct {
			Scope struct {
				Type  string `json:"type"`
				Token string `json:"token"`
			} `json:"scope"`
		} `json:"payload"`
	} `json:"directive"`
}

// These two structures handle the "Properies" field of the Discovery
// response.

type SupportedResponse struct {
	Name string `json:"name"`
}

type PropertiesResponse struct {
	Supported   []SupportedResponse `json:"supported,omitempty"`
	Proactive   bool                `json:"proactivelyReported,omitempty"`
	Retrievable bool                `json:"retrievable,omitempty"`
}

// Still part of the Discovery process...
//
// In this structure, Properties needs to be a response to allow "omitempty"
// to work properly.  Which is a mess.

type CapabilitiesResponse struct {
	Type       string              `json:"type"`
	Interface  string              `json:"interface"`
	Version    string              `json:"version"`
	Properties *PropertiesResponse `json:"properties,omitempty"`
}

type EndpointResponse struct {
	Capabilities      []CapabilitiesResponse `json:"capabilities"`
	Description       string                 `json:"description"`
	DisplayCategories []string               `json:"displayCategories,omitempty"`
	EndpointID        string                 `json:"endpointId"`
	FriendlyName      string                 `json:"friendlyName"`
	ManufacturerName  string                 `json:"manufacturerName"`
}

// All of that finally lets us define the Discovery Response structure
type DiscoveryResponse struct {
	Event struct {
		Header  HeaderStruct `json:"header"`
		Payload struct {
			Endpoints []EndpointResponse `json:"endpoints"`
		} `json:"payload"`
	} `json:"event"`
}

// And this is the simpler StateResponse stuff.  The fields appear
// to be inconsistent with the Discovery message, even though they
// have the same names.
//
// And the Value field needs to be an interface{} because some responses
// a simple string {"value": "foobar"} but other responses are an embedded
// structure {"value":{"value": "foobar"}}.  No that's not a misread of the
// specs, it's needed.

type StatePropertiesResponse struct {
	NameSpace    string      `json:"namespace"`
	Name         string      `json:"name"`
	Value        interface{} `json:"value"`
	TimeOfSample time.Time   `json:"timeOfSample"`
	Uncertainty  int         `json:"uncertaintyInMilliseconds"`
}

type StateResponse struct {
	Event struct {
		Header   HeaderStruct `json:"header"`
		Endpoint struct {
			EndpointID string `json:"endpointId"`
		} `json:"endpoint"`
		Payload struct {
		} `json:"payload"`
	} `json:"event"`
	Context struct {
		Properties []StatePropertiesResponse `json:"properties"`
	} `json:"context"`
}

// How to push a ChangeReport

type ChangeReport struct {
	Event struct {
		Header   HeaderStruct `json:"header"`
		Endpoint struct {
			Scope struct {
				Type  string `json:"type"`
				Token string `json:"token"`
			} `json:"scope"`
			EndpointID string `json:"endpointId"`
		} `json:"endpoint"`
		Payload struct {
			Change struct {
				Cause struct {
					Type string `json:"type"`
				} `json:"cause"`
				Properties []StatePropertiesResponse `json:"properties"`
			} `json:"change"`
		} `json:"payload"`
	} `json:"event"`
	Context struct {
		Properties []StatePropertiesResponse `json:"properties"`
	} `json:"context"`
}

// This is the incoming AcceptGrant format

type GrantRequest struct {
	Directive struct {
		Header  HeaderStruct `json:"header"`
		Payload struct {
			Grant struct {
				Type string `json:"type"`
				Code string `json:"code"`
			} `json:"grant"`
			Grantee struct {
				Type  string `json:"type"`
				Token string `json:"token"`
			} `json:"grantee"`
		} `json:"payload"`
	} `json:"directive"`
}

// AcceptGrant response is simple

type GrantRequestResponse struct {
	Event struct {
		Header  HeaderStruct `json:"header"`
		Payload struct {
			Type    string `json:"type,omitempty"`
			Message string `json:"message,omitempty"`
		} `json:"payload"`
	} `json:"event"`
}

// Structure we get back from Auth endpoint
// We also add a timestamp when we set it
type AuthResponse struct {
	Error        string `json:"error"`
	ErrorDesc    string `json:"error_description"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TimeSet      int64
}
