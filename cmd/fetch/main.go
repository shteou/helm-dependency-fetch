package main

import (
	"fmt"
	"log"

	"github.com/shteou/helm-dependency-fetch/pkg/fetch"
	"github.com/spf13/afero"
)

func main() {
	f := fetch.NewHelmDependencyFetch(afero.NewOsFs())

	dependencies, err := f.ParseDependencies()
	if err != nil {
		log.Fatalf("Error parsing dependencies: %+v", err)
	}

	err = f.CreateChartsDirectory()
	if err != nil {
		log.Fatalf("Failed to manage charts directory %+v", err)
	}

	for _, dependency := range *dependencies {
		fmt.Printf("Fetching %s @ %s\n", dependency.Name, dependency.Version)
		err := f.FetchVersion(dependency)
		if err != nil {
			log.Fatalf("Error fetching dependencies: %+v", err)
		}
	}
}
