pulsekit
========

Repository for tools / libraries built around Zutubi Pulse server for automation purposes. This most notable tools in this repository are:

## cmd/pulsecli

Command-line tool for communicating with Zutubi Pulse server via its Remote API.

#### Installation

Just go get it:

```
~ $ go get github.com/x-formation/pulsekit/cmd/pulsecli
~ $ GOBIN=~/bin go install github.com/x-formation/pulsekit/cmd/pulsecli
```

Ensure you have `$GOPATH` set and `$GOBIN` (or `$GOPATH`/bin) is in your `$PATH`.

#### Usage

```
NAME:
   pulsecli - a command-line client for a Pulse server

USAGE:
   pulsecli [global options] command [command options] [arguments...]

VERSION:
   0.1.0

COMMANDS:
   login     Creates or updates session for current user
   trigger   Triggers a build
   init      Initialises a project
   health    Performs a health check
   projects  Lists all projct names
   stages    Lists all stage names
   agents    Lists all agent names
   status    Lists build's status
   build     Gives build ID associated with given request ID
   wait   Waits for a build to complete
   personal  Sends a personal build request
   help, h Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --addr 'http://pulse/xmlrpc'	Pulse Remote API endpoint
   --user ''              Pulse user name
   --pass ''              Pulse user password
   --agent, -a '.*'       Agent name patter
   --project, -p '.*'     Project name pattern (or "personal")
   --stage, -s '.*'       Stage name pattern
   --timeout, -t '15s'    Maximum wait time
   --patch ''             Patch file for a personal build
   --revision, -r 'HEAD'  Revision to use for personal build
   --build, -b '0'        Build number
   --prtg                 PRTG-friendly output
   --version, -v          print the version
   --help, -h             show help
```

#### PRTG output

Passing `--prtg` flag makes the output PRTG-friendly - when command exists with:

* exit code 0, the output is:

`0:0:OK`

* exit code different than 0, the output is:

`2:1:"<error message here>"`

#### Examples

###### Store credentials in `$HOME`

Credentials are stored for a current user in `~/.pulsecli`. They're stored as
plain text, so be careful if others have access to your `$HOME` directory.

```
~ $ pulsecli --user $USER --pass $PASS login
~ $ pulsecli --prtg health
0:0:OK
```

###### Perform a health check against Pulse server

```
~ $ pulsecli --prtg health
0:0:OK
```

###### Perform a health check against `LM-X - Tier 1` project

The output is in the YAML format.

```
~ $ pulsecli -p 'LM-X - Tier 1' health
LM-X - Tier 1 (build 1356):
- severity: error
  message: Recipe LM-X Unix failed
  stagename: ""
  commandname: ""
  artifactname: ""
  path: ""
- severity: error
  message: Recipe LM-X Windows failed
  stagename: ""
  commandname: ""
  artifactname: ""
  path: ""
- severity: error
  message: Command 'Run unittest_licserver' failed
  stagename: Build - Windows x86
  commandname: ""
  artifactname: ""
  path: ""
...
```

###### Trigger builds for all the projects except `LAC`

`trigger` outputs a list of `<request id>	<project name>` pairs, separated
by a tab. In order to obtain the build ID run `pulsecli build <request id>`.

```
~ $ pulsecli --project '^((\S)[^C]+|Pulse CLI.*)$' trigger
2242787	"License Statistics - Development"
2243158	"LM-X - Tier 1"
2243466	"LM-X - Tier 2"
2243871	"Sales Assistant - Deployment"
2244123	"LM-X - Deployment"
...
2248358	"Go - Database"
2248525	"Sales Assistant - Tests and Installer"
2248659	"Shared PHP Library"
2248803	"Go - Unittest"
2248953	"Sales Assistant - deploy"
2249091	"Go - Accept (devel)"
2249227	"LM-X - Solaris 10 test"
2249380	"Puppet Node Tests"
```

###### Wait for the build triggered by request ID `2248358` to complete

```
~ $ pulsecli -p 'Go - Database' -b 2248358 wait
```

###### Trigger multiple projects and wait for each to complete

```
~ $ pulsecli -p 'Pulse CLI' trigger | xargs printf -- '-b %d -p \"%s\"\n' | parallel -- eval "pulsecli -t 1m {} wait"
```

###### Request a personal build for `review-1234.diff` and `Pulse CLI` project

```
~ $ git diff HEAD~1 > review-1234.diff
~ $ pulsecli -p 'Pulse CLI' --patch review-1234.diff personal
542
```

###### Request a personal build for stages `Build - MAC OS X 10.9` and `Build - Linux x86`

```
~ $ git diff HEAD~1 > review-4321.diff
~ $ pulsecli -p 'LM-X - Tier 1' --patch review-4321.diff --stage 'Build - (MAC OS X|Linux x86$)' personal
207
```

###### Obtain a build ID for the `2260289` request ID

```
~ $ pulsecli build 2260289
130
```

###### Trigger a build for all `LM-X` tiers

```
~ $ pulsecli --project 'LM-X - Tier' trigger
2238515	"LM-X - Tier 1"
2238845	"LM-X - Tier 2"
```

###### Initialise all projects within `Pulse CLI` group

```
~ $ pulsecli --project 'Pulse CLI' init
true	"Pulse CLI - Failure"
true	"Pulse CLI"
true	"Pulse CLI - Success"
```

###### List all the projects

```
~ $ pulsecli projects
License Statistics - Development
C++ - Unittest
LM-X - Deployment
Go - Unittest (devel)
...
C++ - Database
Go - Database
Go - Unittest
Clang Profiling
LM-X - Solaris 10 test
License Activation Center - Accept SOAP
License Activation Center - Accept REST
```

###### List all the stages for the `License Activation Center - API` project

```
~ $ pulsecli --project 'License Activation Center - API' stages
Build - VC2010 - API
Build - Linux x86 - API
Build - VC2010 x64 - API
Build - Linux x64 - API
Build - Mac OSX Universal 10.9 - API
```

###### List all the agents

`agents` outputs a list of `<agent hostname>	<agent name>` pairs, separated
by a tab. It tries to parse an agent's URL and output a hostname only. When
parsing fails (e.g. Pulse 2.6.19 does not put ipv6 addresses in brackets)
it outputs an URL instead.

```
~ $ pulsecli agents
aix275	 "AIX - 5.3"
freebsd10_x64	 "FreeBSD 10 - x64"
hpuxia64	 "HPUX - IA64"
pulse-arm	 "Linux - ARM"
centos5_x64	 "Linux - CentOS 5.10 - Distrib - x64"
coverage	 "Linux - Coverage"
pulse-deb50-x64	 "Linux - Debian 5.0 - Distrib - x64"
pulse-deb50-x86	 "Linux - Debian 5.0 - Distrib - x86"
http://8000:8000:8000:8000:250:56ff:febc:619a:8090	 "Linux - IPv6"
...
pulse-win-7	 "Windows 8.1 - 7"
pulse-win-8	 "Windows 8.1 - 8"
pulse-win-9	 "Windows 8.1 - 9"
```

###### Get a status of the `LM-X - Release Build - Tier 2` project

The output is in the YAML format.

The `--build` or `-b` flag expects either:

  * a real build number
  * 0 which means latest build number
```
~ $ pulsecli -p 'LM-X - Release Build - Tier 2' -b 0 status
LM-X - Release Build - Tier 2 (build 547):
- id: 547
  complete: true
  end: {}
  endunix: "1396963267179"
  errors: 0
  maturity: integration
  owner: LM-X - Release Build - Tier 2
  personal: false
  pinned: false
  progress: -1
  project: LM-X - Release Build - Tier 2
  revision: 887e88a5c4709e9bf260744d398d71dd7ef70050
  reason: manual trigger by rjeczalik
...
```
  * a negative number being an relative offset to the latest build number
```
~ $ pulsecli -p 'LM-X - Release Build - Tier 2' -b -10 status
LM-X - Release Build - Tier 2 (build 537):
- id: 537
  complete: true
  end: {}
  endunix: "1396372083402"
  errors: 8
  maturity: integration
  owner: LM-X - Release Build - Tier 2
  personal: false
  pinned: false
  progress: -1
  project: LM-X - Release Build - Tier 2
  revision: 22fc614bd290041778ad1a69fc66c97841c77177
...
```
