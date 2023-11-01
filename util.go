package main

import (
	"fmt"
	"log"
	"os"
)

func getGithubClientID() (githubClientID string) {
	githubClientID, ok := os.LookupEnv("CLIENT_ID")
	if !ok {
		log.Fatal("github client id not defined in .env file")
	}
	return
}

func getGithubClientSecret() (githubClientSecret string) {
	githubClientSecret, ok := os.LookupEnv("CLIENT_SECRET")
	if !ok {
		log.Fatal("github client secret not defined in .env file")
	}
	return
}

func logger(any ...any) {
	fmt.Println("\n", any)
}
