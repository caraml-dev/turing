# Vertagen 

version + tag + generator - tiny project that helps to generate a semantic version 
of your project artefacts (docker images, library packages, etc.) from the git tag.
---

### How it works

If run inside a git repository, this script will return a valid semver based on 
the semver formatted tags. 

For example if the current HEAD is tagged at `0.0.1`, then the version echoed will 
simply be `0.0.1`. However, if the tag is say, 3 commits behind, the tag will be in 
the form `0.0.1-build.3-0ace960`. This is basically, the current tag a monotonically 
increasing commit (the number of commits since the tag), and then a git short ref 
to identify the commit.

It's also possible to generate a semver, that is based on the git tags with a 
specific prefix. This can be useful for monorepos, where different components inside
the repo can have different version numbers. For example if we want to get a semver 
from tags with `sdk/` prefix (i.e. `sdk/v0.0.1`), then we can use run vertagen with a 
`--prefix` argument:
```shell
vertagen --prefix sdk/
```
---

### Usage
```
$ vertagen <options>
    -h, --help               Display help
    -p, --prefix             (Optional) Prefix used to filter git tags. Default: (empty)
    -b, --main-branch        (Optional) Repository's main branch. Default: 'main'
    -y, --pypi-compatible    Generate PyPi compatible version. Default: (false)
```

### Fun Fact
ver·ta·gen /fɛɐ̯ˈtaːɡn̩,vertágen/ – (German)
verb
    1. postpone
    2. adjourn
    3. defer
    4. prorogue
    5. procrastinate
