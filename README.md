
# rsyncy

A status/progress bar for [rsync](https://github.com/WayneD/rsync).

![gif of rsyncy -a a/ b](https://raw.githubusercontent.com/wiki/laktak/rsyncy/readme/demo.gif "rsyncy -a a/ b")


- [Status Bar](#status-bar)
- [Usage](#usage)
- [Installation](#installation)
- [Known Issue](#known-issue-when-using-ssh-behind-rsync)
- [lf (TUI) support](#lf-tui-support)
- [Development](#development)


## Status Bar

```
[#######:::::::::::::::::::::::]  25% |     100.60M |     205.13kB/s | 0:00:22 | #3019 | 69% (4422..)
[########################::::::]  82% |     367.57M |     508.23kB/s | 0:00:44 | #4234 | 85% of 5055 files
```

The status bar shows the following information:

Description | Sample
--- | ---
(1) Progress bar with percentage of the total transfer | `[########################::::::]  80%`
(2) Bytes transferred | `19.17G`
(3) Transfer speed | `86.65MB/s`
(4) Elapsed time since starting rsync | `0:03:18`
(5) Number of files transferred | `#306`
(6) Files<br>- percentage completed<br>- `*` spinner and `..` are shown while rsync is still scanning | `69% (4422..) *`<br>`85% of 5055 files`

The spinner indicates that rsync is still looking for files. Until this process completes the progress bar may decrease as new files are found.


## Usage

`rsyncy` is a wrapper around `rsync`.

- You run `rsyncy` with the same arguments as it will pass them to `rsync` internally.
- You do not need to specify any `--info` arguments as rsyncy will add them automatically (`--info=progress2 -hv`).

```
# simple example
$ rsyncy -a FROM/ TO
```

Alternatively you can pipe the output from rsync to rsyncy (in which case you need to specify `--info=progress2 -hv` yourself).

```
$ rsync -a --info=progress2 -hv FROM/ TO | rsyncy
```

At the moment `rsyncy` itself has only one option, you can turn off colors via the `NO_COLOR=1` environment variable.


## Installation

rsync is implemented in Go. For legacy reasons there is also a Python implementation that is still maintained. Both versions should behave exactly the same.


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
$ go install github.com/laktak/rsyncy/v2@latest
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
```



## Known Issue when using ssh behind rsync

ssh uses direct TTY access to make sure that the input is indeed issued by an interactive keyboard user (for host keys and passwords). That means that rsyncy does not know that ssh is waiting for input and will draw the status bar over it. You can still enter your password and press enter to continue.

Workaround: connect once to your server via ssh to add it to the known_hosts file.


## lf (TUI) support

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

