# github-release

> [!WARNING]  
> `github-release` is deprecated. We recommend creating GitHub releases with [the `gh` CLI tool](https://github.com/cli/cli).

`github-release` is a utility to create GitHub releases and upload packages.

### Usage

```bash
$ github-release <release name> <fileglob> \
    --target <target> \ # defaults to master, for release commitish
    --commit <sha> \ # commit hash for tag ref
    --tag <tag> \ # defaults to the name of the release
    --prerelease \ # defaults to false
    --update \ # update if release exists, defaults to false
    --github-repository <userorg/repo> \
    --github-access-token <token>
```

Environment variables can also be used:

```bash
$ export GITHUB_RELEASE_ACCESS_TOKEN="..."
$ export GITHUB_RELEASE_REPOSITORY="..."
$ export GITHUB_RELEASE_TAG="..."
$ export GITHUB_RELEASE_TARGET="..."
$ export GITHUB_RELEASE_COMMIT="..."
$ export GITHUB_RELEASE_PRERELEASE="..."
$ export GITHUB_RELEASE_UPDATE="..."
$ github-release "v1.0" pkg/*.tar.gz
```

For the GitHub access token, you can use a [personal access token](https://github.com/settings/applications#personal-access-tokens)

### Development

```
git clone git@github.com:buildkite/github-release.git
cd github-release
direnv allow
go run main.go --help
```

### Sponsor

This project is developed and maintained by [Buildkite](https://buildkite.com)

### Contributing

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request

### Copyright

Copyright (c) 2015 Keith Pitt, Tim Lucas, Buildkite Pty Ltd. See LICENSE for details.
