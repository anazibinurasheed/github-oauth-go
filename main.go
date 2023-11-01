package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func init() {
	// Loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Fatal("No .env file found")
	}
}

func main() {
	// Simply returns a link to the login route
	http.HandleFunc("/", rootHandler)

	// Login route
	http.HandleFunc("/login/github/", githubLoginHandler)

	// Github callback
	http.HandleFunc("/login/github/callback", githubCallbackHandler)

	// Route where the authenticated user is redirected to
	http.HandleFunc("/logged-in", func(w http.ResponseWriter, r *http.Request) {
		loggedinHandler(w, r, "")
	})

	go fmt.Println("[server is up now ...]")

	log.Panic(
		http.ListenAndServe(":3000", nil),
	)

}

func loggedinHandler(w http.ResponseWriter, r *http.Request, githubData string) {
	if githubData == "" {
		// Unauthorized users get an unauthorized message
		fmt.Fprintf(w, "unauthorized!")
		return
	}

	// Set return type JSON
	w.Header().Set("Content-Type", "application/json")

	// Prettifying the json
	var prettyJSON bytes.Buffer
	logger("githubData->", githubData)

	// json.Indent is a library utility function to prettify JSON indentation
	err := json.Indent(&prettyJSON, []byte(githubData), "", "\t")
	if err != nil {
		log.Panic("JSON parse error")
	}
	logger("prettyJSON->", prettyJSON)

	// Return the prettified JSON as a string
	fmt.Fprintf(w, string(prettyJSON.Bytes()))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `<a href="/login/github/">LOGIN</a>`)
}

func githubLoginHandler(w http.ResponseWriter, r *http.Request) {
	// Get the environment variable
	githubClientID := getGithubClientID()
	callbackURI := "http://localhost:3000/login/github/callback"

	// Create the dynamic redirect URL for login
	redirectURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s",
		githubClientID,
		callbackURI,
	)

	http.Redirect(w, r, redirectURL, 301)
}

// Once the user accepts, a request is sent by Github to the route
func githubCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	githubAccessToken := getGithubAccessToken(code)

	githubData := getGithubData(githubAccessToken)

	loggedinHandler(w, r, githubData)
}

func getGithubAccessToken(code string) string {
	clientID := getGithubClientID()
	clientSecret := getGithubClientSecret()

	requestBodyMap := map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"code":          code,
	}
	requestJSON, _ := json.Marshal(requestBodyMap)

	// POST request to set URL
	req, err := http.NewRequest(
		"POST",
		"https://github.com/login/oauth/access_token",
		bytes.NewBuffer(requestJSON),
	)

	if err != nil {
		log.Panic("request creation failed")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	//Get the response

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Panic("request failed")
	}

	// Response body converted to stringified JSON
	respbody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error while reading the response body")
	}
	// Represents the response received from Github
	type githubAccessTokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}

	// Convert stringified JSON to a struct object of type githubAccessTokenResponse
	var ghResp githubAccessTokenResponse
	json.Unmarshal(respbody, &ghResp)

	//Return the access token (as the rest of the
	//details are relatively unnecessary for us)
	return ghResp.AccessToken
}

func getGithubData(accessToken string) string {
	// GET request to a set URL
	req, err := http.NewRequest(
		"GET",
		"https://api.github.com/user",
		nil,
	)

	if err != nil {
		log.Panic("API request creation failed")
	}

	// Set the Authorization header before sending the request
	// Authorization: token XXXXXXXXXXXXXXXXXXXXXXXXXXX
	authorizationToken := fmt.Sprintf("token %s", accessToken)
	req.Header.Set("Authorization", authorizationToken)

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Panic("request failed")
	}

	// Read the response as a byte slice
	respBody, _ := ioutil.ReadAll(resp.Body)
	
	// Convert byte slice to string and return
	return string(respBody)
}
