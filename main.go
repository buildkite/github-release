package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/google/go-github/github"
	"github.com/oleiade/reflections"
	"golang.org/x/oauth2"
)

var commandLineName = "github-release"

var commandLineVersion = "1.0"

var commandLineUsage = `github-release is a utility to create GitHub releases and upload packages.

Usage:
  $ github-release "v1.0" pkg/*.tar.gz --commit "branch-or-sha" \ # defaults to master
                                       --tag "1-0-0-stable" \ # defaults to the name of the release
                                       --prerelease \ # defaults to false
                                       --github-repository "your/repo" \
                                       --github-access-token [..]

Environment variables can also be used:

  $ export GITHUB_RELEASE_ACCESS_TOKEN="..."
  $ export GITHUB_RELEASE_REPOSITORY="..."
  $ export GITHUB_RELEASE_TAG="..."
  $ export GITHUB_RELEASE_COMMIT="..."
  $ export GITHUB_RELEASE_PRERELEASE="..."
  $ github-release "v1.0" pkg/*.tar.gz

Version:
  $ github-release --version

Help:
  $ github-release --help

See https://github.com/buildkite/github-release and the GitHub
create release documentation https://developer.github.com/v3/repos/releases/#create-a-release
for more information.`

type commandLineOptions struct {
	GithubAccessToken string `flag:"github-access-token" env:"GITHUB_RELEASE_ACCESS_TOKEN" required:"true"`
	GithubRepository  string `flag:"github-repository" env:"GITHUB_RELEASE_REPOSITORY" required:"true"`
	Tag               string `flag:"tag" env:"GITHUB_RELEASE_TAG"`
	Commit            string `flag:"commit" env:"GITHUB_RELEASE_COMMIT"`
	Prerelease        bool   `flag:"prerelease" env:"GITHUB_RELEASE_PRERELEASE"`
	Draft             bool   `flag:"draft" env:"GITHUB_RELEASE_DRAFT"`
}

// tokenSource is an oauth2.TokenSource which returns a static access token
type tokenSource struct {
	token *oauth2.Token
}

// Token implements the oauth2.TokenSource interface
func (t *tokenSource) Token() (*oauth2.Token, error) {
	return t.token, nil
}

func main() {
	if len(os.Args) == 1 {
		exitAndError("invalid usage")
	}

	// Are they checking version?
	if os.Args[1] == "--version" {
		fmt.Printf("%s version %s\n", commandLineName, commandLineVersion)
		os.Exit(0)
	}

	// Collect the release assets from the command line
	releaseAssets := collectReleaseAssets(os.Args[1:])

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
		fieldKind, _ := reflections.GetFieldKind(opts, fieldName)

		if fieldKind == reflect.String {
			flags.String(flagName, "", "")
		} else if fieldKind == reflect.Bool {
			flags.Bool(flagName, false, "")
		} else {
			exitAndError(fmt.Sprintf("Could not create flag for %s", fieldName))
		}
	}

	// Define our custom usage text
	flags.Usage = func() {
		fmt.Printf("%s\n", commandLineUsage)
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
		fieldKind, _ := reflections.GetFieldKind(opts, fieldName)

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

		// Set the value in the right way
		if fieldKind == reflect.String {
			reflections.SetField(opts, fieldName, value)
		} else if fieldKind == reflect.Bool {
			// The bool is converted to a string above
			if value == "true" {
				reflections.SetField(opts, fieldName, true)
			} else {
				reflections.SetField(opts, fieldName, false)
			}
		} else {
			exitAndError(fmt.Sprintf("Could not set value of %s", fieldName))
		}
	}
}

func exitAndError(message interface{}) {
	fmt.Printf("%s: %s\nSee '%s --help'\n", commandLineName, message, commandLineName)
	os.Exit(1)
}

func release(releaseName string, releaseAssets []string, options *commandLineOptions) {
	if options.Prerelease {
		log.Printf("Creating prerelease %s for repository: %s", releaseName, options.GithubRepository)
	} else {
		log.Printf("Creating release %s for repository: %s", releaseName, options.GithubRepository)
	}

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

	// log.Printf("%s", repos)
	// log.Printf("name: %s", releaseName)
	// log.Printf("assets: %s", releaseAssets)
	// log.Printf("commit: %s", options.Commit)
	// log.Printf("tag: %s", tag)

	// Create an oAuth Token Source
	ts := &tokenSource{
		&oauth2.Token{AccessToken: options.GithubAccessToken},
	}

	// New oAuth client
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	// Github Client
	client := github.NewClient(tc)

	ctx := context.TODO()

	// Create an object that represents the release
	release := &github.RepositoryRelease{
		Name:            &releaseName,
		TagName:         &tagName,
		TargetCommitish: &options.Commit,
		Prerelease:      &options.Prerelease,
    Draft:           &options.Draft,
	}

	// Create the GitHub release
	createdRelease, _, err := client.Repositories.CreateRelease(ctx, repositoryParts[0], repositoryParts[1], release)
	if err != nil {
		log.Fatalf("Failed to create release (%T %v)", err, err)
	}

	// log.Printf("DEBUG: %s", github.Stringify(createdRelease))

	// Start uploading the assets
	for i := 0; i < len(releaseAssets); i++ {
		fileName := releaseAssets[i]

		file, err := os.Open(fileName)
		if err != nil {
			log.Fatalf("Could not open file \"%s\" (%T %v)", fileName, err, err)
		}

		releaseAssetOptions := &github.UploadOptions{Name: filepath.Base(fileName)}
		createdReleaseAsset, _, err := client.Repositories.UploadReleaseAsset(ctx,repositoryParts[0], repositoryParts[1], *createdRelease.ID, releaseAssetOptions, file)
		if err != nil {
			log.Fatalf("Failed to upload asset \"%s\" (%T %v)", fileName, err, err)
		}

		log.Printf("Successfully uploaded asset: %s", github.Stringify(createdReleaseAsset.URL))
	}

	log.Printf("Successfully created release: %s", github.Stringify(createdRelease.HTMLURL))
}
