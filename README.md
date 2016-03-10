# slatic
slatic is convert slack messages to json file.

## Installation

Download binary file.

- [slatic_0.1.1_darwin_amd64.zip](https://github.com/zaru/slatic/releases/download/v0.1.1/slatic_0.1.1_darwin_amd64.zip)

## Usage

Perform the initial setting. In your home directory `~/.slatic` are created.

If the token is not, it will generate [here](https://api.slack.com/docs/oauth-test-tokens).

```
$ ./slatic init
[Slack token] > (input your token)
[Select channel]
#random
#starwars
please input channel name > (input name)
```

It outputs the previous day of message data in the JSON file.

```
$ ./slatic
$ cat 2016-3-9.json
```
