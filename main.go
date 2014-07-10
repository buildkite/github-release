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

var commandLineUsage = `github-release is a utility to create GitHub releases and upload packages.

Usage:
  $ github-release "v1.0" pkg/*.tar.gz --commit "branch-or-sha" \ # defaults to master
                                       --tag "1-0-0-stable" \ # defaults to the name of the release
                                       --prerelease "true" \ # defaults to false
                                       --github-repository "your/repo" \
                                       --github-access-token [..]

Environment variables can also be used:

  $ export GITHUB_RELEASE_ACCESS_TOKEN="..."
  $ export GITHUB_RELEASE_REPOSITORY="..."
  $ export GITHUB_RELEASE_TAG="..."
  $ export GITHUB_RELEASE_COMMIT="..."
  $ export GITHUB_RELEASE_PRERELEASE="..."
  $ github-release "v1.0" pkg/*.tar.gz

Help:
  $ github-release --help

See https://github.com/buildboxhq/github-release and the GitHub
create release documentation https://developer.github.com/v3/repos/releases/#create-a-release
for more information.`

type commandLineOptions struct {
	GithubAccessToken string `flag:"github-access-token" env:"GITHUB_RELEASE_ACCESS_TOKEN" required:"true"`
	GithubRepository  string `flag:"github-repository" env:"GITHUB_RELEASE_REPOSITORY" required:"true"`
	Tag               string `flag:"tag" env:"GITHUB_RELEASE_TAG"`
	Commit            string `flag:"commit" env:"GITHUB_RELEASE_COMMIT"`
	Prerelease        string `flag:"prerelease" env:"GITHUB_RELEASE_PRERELEASE"`
}

func main() {
	if len(os.Args) == 1 {
		exitAndError("invalid usage")
	}

	// Collect the release assets from the command line
	releaseAssets := collectReleaseAssets(os.Args)

	// Parse our command line options
	options := commandLineOptions{}
	parseArgs(&options, os.Args)

	// Grab our release name. If it starts with a '--', then they haven't
	// entered one.
	releaseName := os.Args[1]
	if strings.HasPrefix(releaseName, "--") {
		exitAndError("invalid usage")
	}

	// fmt.Printf("name: %s\n", releaseName)
	// fmt.Printf("assets: %s\n", releaseAssets)
	// fmt.Printf("options: %s\n", options.GithubAccessToken)

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

// Reflects on the values in the commandLineOptions structure, and fills it
// with values either from the command line, or from the ENV.
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

	// Define our custom usage text
	flags.Usage = func() {
		fmt.Printf("%s\n", commandLineUsage)
		os.Exit(0)
	}

	// Search the args until we find a "--", which signifies the user has started
	// defining options.
	var argumentFlags []string
	started := false
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "--") {
			started = true
		}

		if started {
			argumentFlags = append(argumentFlags, args[i])
		}
	}

	// Now parse the flags
	flags.Parse(argumentFlags)

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
	log.Printf("Creating release %s for repository: %s", releaseName, options.GithubRepository)

	// Split the repository into two parts (owner and repository)
	repositoryParts := strings.Split(options.GithubRepository, "/")
	if len(repositoryParts) != 2 {
		exitAndError("github-repository is in the wrong format")
	}

	// If no tag was provided, use the name of the release
	tagName := options.Tag
	if tagName == "" {
		tagName = releaseName
	}

	// Toggle prerelease
	prerelease := false
	if options.Prerelease == "true" {
		prerelease = true
	}

	// log.Printf("%s", repos)
	// log.Printf("name: %s", releaseName)
	// log.Printf("assets: %s", releaseAssets)
	// log.Printf("commit: %s", options.Commit)
	// log.Printf("tag: %s", tag)

	// Create an oAuth transport
	t := &oauth.Transport{
		Token: &oauth.Token{AccessToken: options.GithubAccessToken},
	}

	// Create a GitHub client with the transport
	client := github.NewClient(t.Client())

	// Create an object that represents the release
	release := &github.RepositoryRelease{
		Name:            &releaseName,
		TagName:         &tagName,
		TargetCommitish: &options.Commit,
		Prerelease:      &prerelease,
	}

	// Create the GitHub release
	createdRelease, _, err := client.Repositories.CreateRelease(repositoryParts[0], repositoryParts[1], release)
	if err != nil {
		log.Fatalf("Failed to create release: %T %v", err, err)
	}

	log.Printf("Successfully created release: %s", github.Stringify(createdRelease.HTMLURL))
}
