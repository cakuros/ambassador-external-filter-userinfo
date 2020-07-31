package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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
	configURL, _ := url.Parse(r.Header.Get("Authorization-URL") + "/.well-known/openid-configuration")
	oidcReq, err := http.NewRequest("GET", configURL.String(), nil)
	if err != nil {
		return nil, err
	}
	oidcRes, err := http.DefaultClient.Do(oidcReq)
	if err != nil {
		return nil, err
	}
	defer oidcRes.Body.Close()
	textBody, err := ioutil.ReadAll(oidcRes.Body)
	if err != nil {
		return nil, err
	}

	// Setup decoder
	//dec := json.NewDecoder(oidcRes.Body)
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

	return &discovery, nil
}

// Send get request with bearer token to URL, returns strmap of response JSON
func httpGet(r *http.Request, url *url.URL) (*map[string]interface{}, error) {
	req, err := http.NewRequest("GET", url.String(), nil)
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
	// Close the response body
	defer res.Body.Close()
	resText, err := ioutil.ReadAll(res.Body)

	var resMap map[string]interface{}
	err = json.Unmarshal([]byte(resText), &resMap)

	return &resMap, nil
}

// Handler function that is called on receiving a request
func handler(w http.ResponseWriter, r *http.Request) {
	// OIDC Discovery
	config, err := Discover(r)
	if err != nil {
		// Handle error
		fmt.Fprintf(w, "%s", err.Error())
	}

	// Config is the results of OIDC Discovery, in this case we want the User Info endpoint
	// Request r is needed to get the Authorization header to submit auth_token
	response, err := httpGet(r, config.UserInfoEndpoint)
	if err != nil {
		// Handle error
		fmt.Fprintf(w, "%s", err.Error())
	}

	// Now we have UserInfo as map[string]interface, need to convert id_token to m[s]i
	fmt.Fprintf(w, "%+v %+v", config, response) // debug

}

func main() {
	// Register the handler function to match pattern '/'
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
