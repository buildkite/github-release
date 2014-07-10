# github-release

github-release is a utility to create GitHub releases and upload packages.

### Usage

```bash
$ github-release "v1.0" pkg/*.tar.gz --commit "branch-or-sha" \
                                     --tag "1-0-0-stable" \
                                     --prerelease "true" \
                                     --github-repository "your/repo" \
                                     --github-access-token [..]
```

Environment variables can also be used:

```bash
$ export GITHUB_RELEASE_ACCESS_TOKEN="..."
$ export GITHUB_RELEASE_REPOSITORY="..."
$ export GITHUB_RELEASE_TAG="..."
$ export GITHUB_RELEASE_COMMIT="..."
$ export GITHUB_RELEASE_PRERELEASE="..."
$ github-release "v1.0" pkg/*.tar.gz
```

For the GitHub access token, you can use a [personal access token](https://github.com/settings/applications#personal-access-tokens)
