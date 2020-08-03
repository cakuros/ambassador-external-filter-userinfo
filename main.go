package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

type oidConfig struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserInfoEndpoint      string `json:"userinfo_endpoint"`
	EndSessionEndpoint    string `json:"end_session_endpoint"`
	JSONWebKeySetURI      string `json:"jwks_uri"`
}

type Discovered struct {
	Issuer                string
	AuthorizationEndpoint *url.URL
	TokenEndpoint         *url.URL
	UserInfoEndpoint      *url.URL
	EndSessionEndpoint    *url.URL
	JSONWebKeySetEndpoint *url.URL
}

func Discover(r *http.Request) (*Discovered, error) {
	fmt.Printf("Starting Discovery\n\n")
	// Pull Env Variable
	fmt.Printf("Pulling env variable\n\n")
	env := os.Getenv("OIDC_SERVER")
	if env == "" {
		fmt.Printf("Missing environment variable OIDC_SERVER")
	}
	// Use environment variable to construct discovery endpoint url
	configURL, _ := url.Parse(env + "/.well-known/openid-configuration")
	fmt.Printf("Discovery URL: %s\n\n", configURL.String())

	// Create new request
	oidcReq, err := http.NewRequest("GET", configURL.String(), nil)
	if err != nil {
		return nil, err
	}

	// Send request and get response
	oidcRes, err := http.DefaultClient.Do(oidcReq)
	if err != nil {
		return nil, err
	}
	defer oidcRes.Body.Close()
	// Read body and save as text
	textBody, err := ioutil.ReadAll(oidcRes.Body)
	if err != nil {
		return nil, err
	}

	fmt.Printf("OIDC Discovery Text: %s\n\n", textBody)

	// Empty struct for OIDC
	var config oidConfig

	err = json.Unmarshal(textBody, &config)
	if err != nil {
		//var syntaxError *json.SyntaxError
		//var unmarshalTypeError *json.UnmarshalTypeError
		return nil, err
	}

	var discovery Discovered
	discovery.Issuer = config.Issuer
	discovery.AuthorizationEndpoint, err = url.Parse(config.AuthorizationEndpoint)
	if err != nil {
		return nil, err
	}
	discovery.TokenEndpoint, err = url.Parse(config.TokenEndpoint)
	if err != nil {
		return nil, err
	}
	discovery.UserInfoEndpoint, err = url.Parse(config.UserInfoEndpoint)
	if err != nil {
		return nil, err
	}
	if config.EndSessionEndpoint != "" {
		discovery.EndSessionEndpoint, err = url.Parse(config.EndSessionEndpoint)
		if err != nil {
			return nil, err
		}
	}
	discovery.JSONWebKeySetEndpoint, err = url.Parse(config.JSONWebKeySetURI)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Discovery struct result: %+v\n\n", discovery)

	return &discovery, nil
}

// Send get request with bearer token to URL, returns strmapinterf of response JSON
func httpPost(r *http.Request, userInfoEndpoint *url.URL) (map[string]interface{}, error) {
	fmt.Printf("Starting POST request to userinfo endpoint with URL: %s\n\n", userInfoEndpoint.String())
	req, err := http.NewRequest("POST", userInfoEndpoint.String(), nil)
	if err != nil {
		// Handle error
		return nil, err
	}
	// Pass on the Authorization header recieved from the request
	req.Header.Set("Authorization", r.Header.Get("Authorization"))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		// Handle error
		return nil, err
	}
	fmt.Printf("Request Sent with Authorization code: %s\n\n", r.Header.Get("Authorization"))
	// Close the response body
	defer res.Body.Close()
	resText, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Text response: %s\n\n", resText)

	var resMap map[string]interface{}
	err = json.Unmarshal([]byte(resText), &resMap)

	if err != nil {
		return nil, err
	}

	fmt.Printf("Map response: %s\n\n", resMap)

	return resMap, nil
}

// Handler function that is called on receiving a request
func handler(w http.ResponseWriter, r *http.Request) {
	// OIDC Discovery
	fmt.Printf("Handler catch, Entering Discovery\n\n")
	config, err := Discover(r)
	if err != nil {
		// Handle error
		w.WriteHeader(http.StatusBadRequest)
	}

	fmt.Printf("Discover returns: %+v, starting UserInfo POST request\n\n", config)
	// fmt.Fprintf(w, "%v\n", config) // debug

	// Config is the results of OIDC Discovery, in this case we want the User Info endpoint
	// Request r is needed to get the Authorization header to submit auth_token
	response, err := httpPost(r, config.UserInfoEndpoint)
	if err != nil || response == nil {
		// Handle error
		w.WriteHeader(http.StatusBadRequest)
	}

	// Now we have UserInfo as map[string]interface, time to extract items!
	id := response["name"] // Set "name" to any requested userinfo and attach it to the response header.
	w.Header().Set("x-userinfo-name", id.(string))
	w.WriteHeader(http.StatusOK)
}

func main() {
	// Register the handler function to match pattern '/'
	http.HandleFunc("/", handler)
	fmt.Printf("Starting ListenAndServe on :8080\n\n")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
