package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/shteou/helm-dependency-fetch/pkg/fetch"
)

func printUsage() {
	fmt.Println("helm-dependency-fetch [chartDir]")
	fmt.Println("  helm-dependency-fetch is a drop in replacement for helm dependency build")
	fmt.Println("  It will fetch the chart dependencies for the supplied chart. If no chart is supplied")
	fmt.Println("  the current directory is assumed to be the chart directory")
	fmt.Println("  Note, lock files are not generated, and are ignored.")
	fmt.Println("  The tool will fetch the latest dependencies on each execution.")
	fmt.Println("")
	fmt.Println("Flags:")
	fmt.Println("  -t, --timeout a timeout, in seconds, for the fetch command")
	fmt.Println("  -h, --help    help for this command")
}

func main() {
	chartDirectory := "."

	flag.Usage = printUsage

	var timeoutSeconds uint
	flag.UintVar(&timeoutSeconds, "timeout", 300, "a timeout, in seconds, for the fetch command")
	flag.UintVar(&timeoutSeconds, "t", 300, "a timeout, in seconds, for the fetch command")

	flag.Parse()
	if flag.NArg() > 0 {
		chartDirectory = flag.Arg(0)
	}

	err := os.Chdir(chartDirectory)
	if err != nil {
		fmt.Printf("Failed to change directory to %s\n", os.Args[1])
		fmt.Printf("Does the supplied chart directory exist?\n")
		return
	}

	f := fetch.NewHelmDependencyFetch()

	dependencies, err := f.ParseDependencies()
	if err != nil {
		log.Fatalf("Error parsing dependencies: %+v", err)
	}

	err = f.CreateChartsDirectory()
	if err != nil {
		log.Fatalf("Failed to manage charts directory %+v", err)
	}

	timeout, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeoutSeconds))
	defer cancel()

	for _, dependency := range *dependencies {
		fmt.Printf("Fetching %s @ %s\n", dependency.Name, dependency.Version)
		err := f.FetchVersion(timeout, dependency)
		if err != nil {
			log.Fatalf("Error fetching dependencies: %+v", err)
		}
	}
}
