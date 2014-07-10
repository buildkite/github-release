package main

import (
	"code.google.com/p/goauth2/oauth"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/google/go-github/github"
	"log"
	"os"
	"strings"
)

var AppHelpTemplate = `A utility to create GitHub releases and upload packages.

Usage:
  $ {{.Name}} "v1.0" --release-assets "pkg/*.tar.gz" \
                          --release-commit "branch-or-sha" \ # defaults to master
                          --github-repository "your/repo" \
                          --github-access-token [..]

Help:
  $ {{.Name}} --help

See https://github.com/buildboxhq/github-release for more information.
`

var AppName = "github-release"

func main() {
	cli.AppHelpTemplate = AppHelpTemplate

	app := cli.NewApp()
	app.Name = AppName
	app.Version = "0.1"
	app.Action = func(c *cli.Context) {
		// There should be 2 args, the key and the value.
		if len(c.Args()) != 1 {
			exitAndError("missing release name")
		}

		// Grab the release name
		releaseName := c.Args()[0]

		// Get arguments from either the command line or ENV
		githubAccessToken := argumentOrEnv(c, "github-access-token", "GITHUB_RELEASE_ACCESS_TOKEN", true)
		githubRepository := argumentOrEnv(c, "github-repository", "GITHUB_RELEASE_REPOSITORY", true)
		releaseCommit := argumentOrEnv(c, "releaseCommit", "GITHUB_RELEASE_COMMIT", false)
		releaseAssets := argumentOrEnv(c, "releaseAssets", "GITHUB_RELEASE_ASSETS", false)

		// Split the repository into two parts (owner and repository)
		repositoryParts := strings.Split(githubRepository, "/")
		if len(repositoryParts) != 2 {
			exitAndError("github-repository is in the wrong format")
		}

		release(githubAccessToken, repositoryParts[0], repositoryParts[1], releaseName, releaseCommit, releaseAssets)
	}

	app.Run(os.Args)
}

func exitAndError(message string) {
	fmt.Printf("%s: %s\nSee '%s --help'\n", AppName, message, AppName)
	os.Exit(1)
}

func argumentOrEnv(c *cli.Context, argumentName string, argumentEnv string, required bool) string {
	value := c.String(argumentName)

	if value == "" {
		value = os.Getenv(argumentEnv)
	}

	if required && value == "" {
		exitAndError(fmt.Sprintf("missing %s", argumentName))
	}

	return value
}

func release(githubAccessToken string, githubOwner string, githubRepository string, releaseName string, releaseCommit string, releaseAssets string) {
	// Create an oAuth transport
	t := &oauth.Transport{
		Token: &oauth.Token{AccessToken: githubAccessToken},
	}

	// Create a GitHub client with the transport
	client := github.NewClient(t.Client())

	// List all repositories for the authenticated user
	repos, _, err := client.Repositories.ListReleases(githubOwner, githubRepository, nil)
	if err != nil {
		log.Fatalf("Failed to get repos: %s", err)
	}

	log.Printf("%s", repos)
	log.Printf("name: %s", releaseName)
	log.Printf("commit: %s", releaseCommit)
	log.Printf("assets: %s", releaseAssets)
}
