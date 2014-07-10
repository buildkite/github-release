# github-release

A command line utility to create GitHub releases and upload packages

### Usage

```bash
$ github-release "v1.0" "pkg/*.tar.gz" --github-repository "your/repo" --github-access-token [..]
```

You can also pass through `github-repository` and `github-access-token` and ENV variables:

```bash
export GITHUB_RELEASE_ACCESS_TOKEN="..."
export GITHUB_RELEASE_REPOSITORY="your/repo"

$ github-release "v1.0" "pkg/*.tar.gz"
```

For the GitHub access token, you can use a [personal access token](https://github.com/settings/applications#personal-access-tokens)
