package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/.well-known/terraform.json", TerraformWellKnownHandler)
	r.HandleFunc("/{namespace}/{name}/{provider}/versions", ListModuleVersionsHandler)
	r.HandleFunc("/{namespace}/{name}/{provider}/{system}/{version}/download", DownloadModuleHandler)

	http.ListenAndServe(":8080", r)
}

func TerraformWellKnownHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `{"modules.v1":"/"}`)
}

func ListModuleVersionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	namespace := vars["namespace"]
	name := vars["name"]
	provider := vars["provider"]

	fmt.Println("namespace", namespace)
	fmt.Println("name", name)
	fmt.Println("provider", provider)

	fmt.Fprint(w, "[]")
}

func DownloadModuleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	namespace := vars["namespace"]
	name := vars["name"]
	provider := vars["provider"]
	system := vars["system"]
	version := vars["version"]

	fmt.Println("namespace", namespace)
	fmt.Println("name", name)
	fmt.Println("provider", provider)
	fmt.Println("system", system)
	fmt.Println("version", version)

	w.WriteHeader(204)
	w.Header().Add("X-Terraform-Get", "/")
}
