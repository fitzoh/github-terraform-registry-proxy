package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"strings"
)

//Required service discovery endpoint
//see https://www.terraform.io/docs/internals/module-registry-protocol.html#service-discovery
const terraformWellKnownEndpoint = "/.well-known/terraform.json"

//Lists the versions available for a given module
//see https://www.terraform.io/docs/internals/module-registry-protocol.html#list-available-versions-for-a-specific-module
//sample URL: http://localhost:8080/terraform-aws-modules/vpc/aws/versions
const listModuleVersionsEndpoint = "/{namespace}/{name}/{provider}/versions"

//Returns the URL to download the source code of a specific module version
//see https://www.terraform.io/docs/internals/module-registry-protocol.html#download-source-code-for-a-specific-module-version
//but I'm pretty sure those docs are wrong, so also see https://github.com/hashicorp/terraform/pull/25964
//sample URL: http://localhost:8080/terraform-aws-modules/vpc/aws/v2.44.0/download
const downloadModuleVersionEndpoint = "/{namespace}/{name}/{provider}/{version}/download"

//Returns the actual content of module
//We need to proxy to it to enable authenticated access
const downloadModuleVersionForRealEndpoint = "/{namespace}/{name}/{provider}/{version}/download-for-real"

func main() {
	r := mux.NewRouter()

	r.HandleFunc(terraformWellKnownEndpoint, terraformWellKnownHandler)
	r.HandleFunc(listModuleVersionsEndpoint, listModuleVersionsHandler)
	r.HandleFunc(downloadModuleVersionEndpoint, downloadModuleHandler)
	r.HandleFunc(downloadModuleVersionForRealEndpoint, downloadModuleForRealHandler)

	r.Use(loggingMiddleware)
	log.Fatal(http.ListenAndServeTLS(":8080", "127.0.0.1.pem", "127.0.0.1-key.pem", r))
}

//Log the request URI and move along
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

//Returns the terraform service discovery payload.
//Currently only the modules API is supported
func terraformWellKnownHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	fmt.Fprint(w, `{"modules.v1":"/"}`)
}

//Gross JSON structs used by listModuleVersionsHandler
type modules struct {
	Modules []versions `json:"modules"`
}
type versions struct {
	Versions []version `json:"versions"`
}
type version struct {
	Version string `json:"version"`
}

//Lists the versions available for a given module
//see https://www.terraform.io/docs/internals/module-registry-protocol.html#list-available-versions-for-a-specific-module
func listModuleVersionsHandler(w http.ResponseWriter, r *http.Request) {
	gh := clientForRequest(r)
	vars := mux.Vars(r)

	namespace := vars["namespace"]
	owner := namespace
	name := vars["name"]
	provider := vars["provider"]

	repo := fmt.Sprintf("terraform-%s-%s", provider, name)

	tags, _, err := gh.Repositories.ListTags(context.Background(), owner, repo, nil)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprint(w, err.Error())
		return
	}

	var vs []version
	for _, tag := range tags {
		vs = append(vs, tagToVersion(tag))
	}

	response := modules{
		Modules: []versions{
			{Versions: vs},
		},
	}
	json.NewEncoder(w).Encode(response)
}

//Simple helper function to map from the github API to terraform structs
func tagToVersion(tag *github.RepositoryTag) version {
	name := tag.GetName()
	if strings.HasPrefix(name, "v") {
		name = name[1:]
	}
	return version{Version: name}
}

//Returns the proxy URL to download the module
func downloadModuleHandler(w http.ResponseWriter, r *http.Request) {
	final := r.URL.String() + "-for-real?archive=tar.gz"
	fmt.Println(final)
	w.Header().Add("X-Terraform-Get", final)
	w.WriteHeader(204)
}

//Finds the tarball URL for the given module tag and proxies an authenticated request for it to the github API
func downloadModuleForRealHandler(w http.ResponseWriter, r *http.Request) {
	gh := clientForRequest(r)
	vars := mux.Vars(r)

	namespace := vars["namespace"]
	name := vars["name"]
	provider := vars["provider"]
	version := vars["version"]

	owner := namespace
	repo := fmt.Sprintf("terraform-%s-%s", provider, name)

	tags, _, err := gh.Repositories.ListTags(context.Background(), owner, repo, nil)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprint(w, err.Error())
		return
	}

	for _, tag := range tags {
		//Terraform strips the leading `v` for semver tags
		if tag.GetName() == version || tag.GetName() == "v"+version {

			fmt.Println("proxying", tag.GetTarballURL())
			downloadReq, _ := http.NewRequest("GET", tag.GetTarballURL(), nil)
			gh.Do(context.Background(), downloadReq, w)
			return
		}
	}
	w.WriteHeader(404)

}

var unauthenticatedGithubClient = github.NewClient(nil)

//Checks if the incoming request contains an access token.
//If it does, return a github client that uses that token.
//Otherwise, return an unauthenticated github client.
func clientForRequest(r *http.Request) *github.Client {
	auth := r.Header.Get("Authorization")
	token := strings.TrimPrefix(auth, "Bearer ")
	if token == "" {
		return unauthenticatedGithubClient
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	client := oauth2.NewClient(ctx, ts)
	return github.NewClient(client)
}
