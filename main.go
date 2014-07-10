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

  $ {{.Name}} "v1.0" "pkg/*.tar.gz" --github-repository "your/repo" --github-access-token [..]

Use "{{.Name}} --help" to see this message.

`

func main() {
	cli.AppHelpTemplate = AppHelpTemplate

	app := cli.NewApp()
	app.Name = "github-release"
	app.Version = "0.1"
	app.Action = func(c *cli.Context) {
		// Get arguments from either the command line or ENV
		githubAccessToken := argumentOrEnv(c, "github-access-token", "GITHUB_RELEASE_ACCESS_TOKEN")
		githubRepository := argumentOrEnv(c, "github-repository", "GITHUB_RELEASE_REPOSITORY")

		// Split the repository into two parts (owner and repository)
		repositoryParts := strings.Split(githubRepository, "/")
		if len(repositoryParts) != 2 {
			fmt.Printf("%s: github-repository is in the wrong format\nSee '%s --help'\n", c.App.Name, c.App.Name)
			os.Exit(1)
		}

		release(githubAccessToken, repositoryParts[0], repositoryParts[1])
	}

	app.Run(os.Args)
}

func argumentOrEnv(c *cli.Context, argumentName string, argumentEnv string) string {
	value := c.String(argumentName)

	if value == "" {
		value = os.Getenv(argumentEnv)
	}

	if value == "" {
		fmt.Printf("%s: missing %s\nSee '%s --help'\n", c.App.Name, argumentName, c.App.Name)
		os.Exit(1)
	}

	return value
}

func release(githubAccessToken string, githubOwner string, githubRepository string) {
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
}
