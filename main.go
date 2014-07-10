package main

import (
	"code.google.com/p/goauth2/oauth"
	"flag"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/oleiade/reflections"
	"log"
	"os"
	"strings"
)

var commandLineName = "github-release"

type commandLineOptions struct {
	GithubAccessToken string `flag:"github-access-token" env:"GITHUB_RELEASE_ACCESS_TOKEN" required:true`
	GithubRepository  string `flag:"github-repository" env:"GITHUB_RELEASE_REPOSITORY" required:true`
	Tag               string `flag:"tag" env:"GITHUB_RELEASE_TAG" required:false`
	Commit            string `flag:"commit" env:"GITHUB_RELEASE_COMMIT" required:false`
}

func main() {
	if len(os.Args) <= 1 {
		exitAndError("invalid usage")
	}

	// Grab our release anme and assets
	releaseName := os.Args[1]
	releaseAssets := collectReleaseAssets(os.Args[2:])

	options := commandLineOptions{}

	// Options will start the argument after the last asset
	parseArgs(&options, os.Args[len(releaseAssets)+2:])

	// fmt.Printf("name: %s\n", releaseName)
	// fmt.Printf("assets: %s\n", releaseAssets)
	// fmt.Printf("options: %s\n", options.GithubAccessToken)

	// If no tag was provided, use the name of the release
	if options.Tag == "" {
		options.Tag = releaseName
	}

	// Finally do the release
	release(releaseName, releaseAssets, &options)
}

// Finds the assets from the argument list by looping over every argument,
// and checking if it's a file.
func collectReleaseAssets(args []string) (files []string) {
	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Check if the file exists
		if _, err := os.Stat(arg); err == nil {
			files = append(files, arg)
		}

		// If the arg is an option, we've gone to far
		if strings.HasPrefix(arg, "--") {
			break
		}
	}

	return
}

func parseArgs(opts *commandLineOptions, args []string) {
	// Create a flag set for args
	flags := flag.NewFlagSet(commandLineName, flag.ExitOnError)

	// Get the fields for the strucutre
	fields, _ := reflections.Fields(opts)

	// Loop through each field of the structure, and define a flag for it
	for i := 0; i < len(fields); i++ {
		fieldName := fields[i]
		flagName, _ := reflections.GetFieldTag(opts, fieldName, "flag")

		flags.String(flagName, "", "")
	}

	// Now parse the flags
	flags.Parse(args)

	// Now the flag set has been populated with values, retrieve them and
	// set them (or use the ENV variable if empty)
	for i := 0; i < len(fields); i++ {
		fieldName := fields[i]

		// Grab the flags we need
		flagName, _ := reflections.GetFieldTag(opts, fieldName, "flag")
		envName, _ := reflections.GetFieldTag(opts, fieldName, "env")
		required, _ := reflections.GetFieldTag(opts, fieldName, "required")

		// Grab the value from the flag
		value := flags.Lookup(flagName).Value.String()

		// If the value of the flag is empty, try the ENV
		if value == "" {
			value = os.Getenv(envName)
		}

		// Do validation
		if required == "true" && value == "" {
			exitAndError(fmt.Sprintf("missing %s", flagName))
		}

		reflections.SetField(opts, fieldName, value)
	}
}

func exitAndError(message interface{}) {
	fmt.Printf("%s: %s\nSee '%s --help'\n", commandLineName, message, commandLineName)
	os.Exit(1)
}

func release(releaseName string, releaseAssets []string, options *commandLineOptions) {
	log.Printf("foo %s", options.GithubRepository)

	// Split the repository into two parts (owner and repository)
	repositoryParts := strings.Split(options.GithubRepository, "/")
	if len(repositoryParts) != 2 {
		exitAndError("github-repository is in the wrong format")
	}

	// Create an oAuth transport
	t := &oauth.Transport{
		Token: &oauth.Token{AccessToken: options.GithubAccessToken},
	}

	// Create a GitHub client with the transport
	client := github.NewClient(t.Client())

	// List all repositories for the authenticated user
	repos, _, err := client.Repositories.ListReleases(repositoryParts[0], repositoryParts[1], nil)
	if err != nil {
		log.Fatalf("Failed to get repos: %s", err)
	}

	log.Printf("%s", repos)
	log.Printf("name: %s", releaseName)
	log.Printf("assets: %s", releaseAssets)
	log.Printf("commit: %s", options.Commit)
	log.Printf("tag: %s", options.Tag)
}
