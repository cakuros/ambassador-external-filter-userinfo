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

func Discover(r *http.Request, authorizationURL string) (*Discovered, error) {
	configURL, _ := url.Parse("https://login.microsoftonline.com/common/v2.0" + "/.well-known/openid-configuration")
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

// Handler function that is called on receiving a request
func handler(w http.ResponseWriter, r *http.Request) {
	// Pull Authorization header
	authHeader := r.Header.Get("Authorization")
	authURLHeader := r.Header.Get("Authorization-URL")

	config, err := Discover(r, authURLHeader)
	if err != nil {
		// Handle error
		fmt.Fprintf(w, "%s", err.Error())
	}

	fmt.Fprintf(w, "%+v %s", config, authHeader)

	// Make call to the IDP
	//	res, err := userInfoGet(r, "{{IDP_URL_ENDPOINT}}", auth_header)
	if err != nil {
		// Handle error
	}
}

func httpGet(r *http.Request, url string, authHeader string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// Handle error
	}
	// Pass on the Authorization header recieved from the request
	req.Header.Add("Authorization", authHeader)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		// Handle error
	}

	// Close the response body
	defer res.Body.Close()

	return res, err
}

func main() {
	// Register the handler function to match pattern '/'
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
