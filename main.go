package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

var gh *github.Client

func main() {
	gh = github.NewClient(nil)
	r := mux.NewRouter()

	r.HandleFunc("/.well-known/terraform.json", TerraformWellKnownHandler)
	//http://localhost:8080/terraform-aws-modules/vpc/aws/versions
	r.HandleFunc("/{namespace}/{name}/{provider}/versions", ListModuleVersionsHandler)
	//http://localhost:8080/terraform-aws-modules/vpc/aws/ignored/v2.44.0/download
	r.HandleFunc("/{namespace}/{name}/{provider}/{system}/{version}/download", DownloadModuleHandler)

	log.Fatal(http.ListenAndServe(":8080", r))
}

func TerraformWellKnownHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	fmt.Fprint(w, `{"modules.v1":"/"}`)
}

type ModulesJson struct {
	Modules []VersionsJson `json:"modules"`
}

type VersionsJson struct {
	Version map[string]string `json:"versions"`
}

func ListModuleVersionsHandler(w http.ResponseWriter, r *http.Request) {
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

	var versions []VersionsJson
	for _, tag := range tags {
		fmt.Println(tag.GetTarballURL())
		fmt.Println(tag.GetZipballURL())
		versions = append(versions, VersionsJson{
			Version: map[string]string{"version": tag.GetName()},
		})
	}

	json.NewEncoder(w).Encode(ModulesJson{
		Modules: versions,
	})
}

func DownloadModuleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	namespace := vars["namespace"]
	name := vars["name"]
	provider := vars["provider"]
	//system is ignored ¯\_(ツ)_/¯
	//system := vars["system"]
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
		if tag.GetName() == version {
			fmt.Println(tag.GetTarballURL())
			w.Header().Add("X-Terraform-Get", tag.GetTarballURL())
			w.WriteHeader(204)

			return
		}
	}
	fmt.Fprint(w, "not found")
	w.WriteHeader(404)

}
