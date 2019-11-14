# (Y)et (A)nother (R)obber: Sail ye seas of git for booty is to be found

<p align="center">
  <img src="https://raw.githubusercontent.com/Furduhlutur/yar/master/images/yargopher3.png" alt="Yar the pirate gopher"/>
</p>

Sail ho! Yar is a tool for plunderin' organizations, users and/or repositories...

In all seriousness though, yar is an OSINT tool for reconnaissance of repositories/users/organizations on Github. Yar clones repositories of users/organizations given to it
and goes through the whole commit history in order of commit time, in search for secrets/tokens/passwords, essentially anything that shouldn't be there. Whenever yar finds a secret,
it will print it out for you to further assess.

Yar searches for secrets either by regex, entropy or both, the choice is yours! Inspired by other git secret grabbers.

## Installation
To install this you simply run the following command.
```
go get github.com/Furduhlutur/yar
```

Just make sure that you have the GOPATH environment variable set in your preferred shell rc and that the $GOPATH/bin directory is in your PATH. More info [here](https://golang.org/doc/code.html#GOPATH).

## Usage
### Want to search for secrets in an organization?
```
yar -o orgname
```

### Want to search for secrets in a users repository?
```
yar -u username
```

### Want to search for secrets in a single repository?
```
yar -r repolink
```
or if you have already cloned the repository
```
yar -r repopath
```

### Want to search for secrets in an organization, for a user and a repository?
```
yar -o orgname -u username -r reponame
```

### Have your own predefined rules?
Rules are stored in a JSON file with the following format:
```
{
    "Rules": [
        {
            "Reason": "The reason for the match",
            "Rule": "The regex rule",
            "Noise": 3
        },
        {
            "Reason": "Super secret token",
            "Rule": "^Token: .*$",
            "Noise": 2
        }
    ]
    "FileBlacklist": [
        "Regex rule here"
        "^.*\\.lock"
    ]
}
```

You can then load your own rule set with the following command:
```
yar -u username --rules PATH_TO_JSON_FILE
```

If you already have a truffleHog config and want to port it over to a yar config there is a script in the config folder that does it for you.
Simply run `python3 trufflestoconfig.py PATH_TO_TRUFFLEHOG_CONFIG` and the script will give you a file named `yarconfig.json`.

### Don't like regex?
```
yar -u username --entropy
```

### Want the best of both worlds?
```
yar -u username --both
```

### Want to search as an authenticated user? 
Add your github token to your environment variables.
```
export YAR_GITHUB_TOKEN=YOUR_TOKEN_HERE
```

### Want to save your findings to a JSON file for later analysis?
```
yar -o orgname --save
```

### Don't like the default colors and want to add your own color settings?
It is possible to customize the colors of the output for Yar through environment variables.
The possible colors to choose from are the following:
```
black
blue
cyan
green
magenta
red
white
yellow
hiBlack
hiBlue
hiCyan
hiGreen
hiMagenta
hiRed
hiWhite
hiYellow
```
Each color can then be suffixed with `bold`, i.e. `blue bold` to make the letters bold.

This is done through the following env variables:
```
YAR_COLOR_VERBOSE -> Color of verbose lines.
YAR_COLOR_SECRET  -> Color of the highlighted secret.
YAR_COLOR_INFO    -> Color of info, that is, simple strings that tell you something.
YAR_COLOR_DATA    -> Color of data, i.e. commit message, reason, etc.
YAR_COLOR_SUCC    -> Color of succesful messages.
YAR_COLOR_WARN    -> Color of warnings.
YAR_COLOR_FAIL    -> Color of fatal warnings.
```
Like so `export YAR_COLOR_SECRET="hiRed bold"`.

## Help
```
usage: yar [-h|--help] [-o|--org "<value>"] [-u|--user "<value>"] [-r|--repo
           "<value>"] [-c|--context <integer>] [-e|--entropy] [-b|--both]
           [-f|--forks] [-n|--noise "<value>"] [--depth <integer>] [--config
           <file>] [--no-bare] [--no-cache] [--no-context] [--include-members]
           [--skip-duplicates] [--cleanup "<value>"] [-s|--save "<value>"]

           Sail ye seas of git for booty is to be found

Arguments:

  -h  --help             Print help information
  -o  --org              Organization to plunder
  -u  --user             User to plunder
  -r  --repo             Repository to plunder
  -c  --context          Show N number of lines for context. Default: 2
  -e  --entropy          Search for secrets using entropy analysis. Default:
                         false
  -b  --both             Search by using both regex and entropy analysis.
                         Overrides entropy flag. Default: false
  -f  --forks            Specifies whether forked repos are included or not.
                         Default: false
  -n  --noise            Specify the range of the noise for rules. Can be
                         specified as up to a certain value (-4), from a
                         certain value (5-), between two values (3-5), just a
                         single value (4) or the whole range (-). Default: -4
      --depth            Specify the depth limit of commits fetched when
                         cloning. Default: 100000
      --config           JSON file containing yar config.
      --no-bare          Clone the whole repository. Default: false
      --no-cache         Don't load from cache. Default: false
      --no-context       Only show the secret itself, similar to trufflehog's
                         regex output. Overrides context flag. Default: false
      --include-members  Include an organization's members for plunderin'.
                         Default: false
      --skip-duplicates  Skip duplicate secrets within repositories. Default:
                         false
      --cleanup          Remove specified cloned directory within yar cache
                         folder. Leave blank to remove the cache folder
                         completely.
  -s  --save             Yar will save all findings to a specified file.
                         Default: findings.json
```

## Acknowledgements
It is important to point out that this idea is inspired by the infamous [truffleHog](https://github.com/dxa4481/truffleHog) tool 
and the code used for entropy searching is in fact borrowed from the truffleHog repository which in turn is borrowed from 
[this blog post](http://blog.dkbza.org/2007/05/scanning-data-for-entropy-anomalies.html).

This project wouldn't have been possible without the following libraries:
+ [go-github](https://github.com/google/go-github/)
+ [go-git](https://github.com/src-d/go-git/)
+ [fatih/color](https://github.com/fatih/color)
