package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// Handler function that is called on receiving a request
func handler(w http.ResponseWriter, r *http.Request) {
	// Pull Authorization header
	auth_header := r.Header.Get("Authorization")

	// Get endpoint, userInfoReq is a JSON blob, key in body = userinfo_endpoint
	userInfoReq, err := http.NewRequest("GET", "https://login.microsoftonline.com/common/v2.0/.well-known/openid-configuration", nil)
	if err != nil {
		// Handle error
	}
	userInfoRes, err := http.DefaultClient.Do(userInfoReq)
	body, err := ioutil.ReadAll(userInfoRes.Body)

	fmt.Fprintf(w, "%s %s", body, auth_header)

	// Make call to the IDP
	//	res, err := userInfoGet(r, "{{IDP_URL_ENDPOINT}}", auth_header)
	if err != nil {
		// Handle error
	}
}

func userInfoGet(r *http.Request, url string, auth_header string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// Handle error
	}
	// Pass on the Authorization header recieved from the request
	req.Header.Add("Authorization", auth_header)
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
