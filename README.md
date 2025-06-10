
# rsyncy

A status/progress bar for [rsync](https://github.com/WayneD/rsync).

![gif of rsyncy -a a/ b](https://raw.githubusercontent.com/laktak/rsyncy/readme/readme/demo-y.gif "rsyncy -a a/ b")


- [Status Bar](#status-bar)
- [Installation](#installation)
- [Usage](#usage)
- [Known Issues when using ssh behind rsync](#known-issues-when-using-ssh-behind-rsync)
- [lf support](#lf-support)
- [Development](#development)


## Status Bar

```
[########################::::::]  80% |      19.17G |      86.65MB/s | 0:03:18 | #306 | scan 46% (2410)\
```

The status bar shows the following information:

Description | Sample
--- | ---
Progress bar with percentage of the total transfer | `[########################::::::]  80%`
Bytes transferred | `19.17G`
Transfer speed | `86.65MB/s`
Elapsed time since starting rsync | `0:03:18`
Number of files transferred | `#306`
Files to scan/check<br>- percentage completed<br>- (number of files)<br>- spinner | `scan 46% (2410)\`

The spinner indicates that rsync is still checking if files need to be updated. Until this process completes the progress bar may decrease as new files are found.


## Installation

### Download Binaries

You can download the official rsyncy binaries for different OS/Platforms from the GitHub releases page. After downloading place it in your `PATH`.

- [github.com/laktak/rsyncy/releases](https://github.com/laktak/rsyncy/releases)

### Install via Homebrew (macOS and Linux)

For macOS and Linux it can also be installed via [Homebrew](https://formulae.brew.sh/formula/rsyncy):

```shell
$ brew install rsyncy
```

### Install via Go

```shell
$ go install github.com/laktak/rsyncy@latest
```

### Install via Pipx

```shell
$ pipx install rsyncy
```

- installs the Python version
- requires [pipx](https://pipx.pypa.io/latest/installation/)

### Build from Source

```shell
$ git clone https://github.com/laktak/rsyncy
$ rsyncy/scripts/build

# binary can be found here
$ ls -l rsyncy/rsyncy
`


## Usage

`rsyncy` is a wrapper around `rsync`.

- You run `rsyncy` with the same arguments as it will pass them to `rsync` internally.
- Do not specify any `--info` arguments, rsyncy will automatically add `--info=progress2` and `-hv` internally.

```
# simple example
$ rsyncy -a FROM/ TO
```

Alternatively you can pipe the output from rsync to rsyncy (in which case you need to specify `--info=progress2 -hv` yourself).

```
$ rsync -a --info=progress2 -hv FROM/ TO | rsyncy
```

At the moment `rsyncy` itself has only one option, you can turn off colors via the `NO_COLOR=1` environment variable.


## Known Issue when using ssh behind rsync

ssh uses direct TTY access to make sure that the input is indeed issued by an interactive keyboard user (for host keys and passwords). That means that rsyncy does not know that ssh is waiting for input and will draw the status bar over it. You can still enter your password and press enter to continue.

Workaround: connect once to your server via ssh to add it to the known_hosts file.


## lf support

`rsyncy-stat` can be used to view only the status output on [lf](https://github.com/gokcehan/lf) (or similar terminal file managers).

Example:

```
cmd paste-rsync %{{
    opt="$@"
    set -- $(cat ~/.local/share/lf/files)
    mode="$1"; shift
    case "$mode" in
        copy) rsyncy-stat -rltphv $opt "$@" . ;;
        move) mv -- "$@" .; lf -remote "send clear" ;;
    esac
}}
```

This shows the copy progress in the `>` line while rsync is running.

If you have downloaded the binary version you can create it with `ln -s rsyncy rsyncy-stat`.


## Development

First record an rsync transfer with [pipevcr](https://github.com/laktak/pipevcr), then replay it to rsyncy when debugging.

