
# rsyncy

A status/progress bar for [rsync](https://github.com/WayneD/rsync).

I love rsync but I always felt it was either too chatty when transferring lots of small files or did not show enough information for the large files in between. rsyncy is a wrapper to change this without having to modify rsync.

![gif of rsyncy -a a/ b](https://raw.githubusercontent.com/laktak/rsyncy/readme/readme/demo-y.gif "rsyncy -a a/ b")

## Status Bar

```
[########################::::::]  80% |      19.17G |      86.65MB/s | 0:03:18 | ETA: 00:00:50 | #306 | scan 46% (2410)\
```

The status bar shows the following information:

Description | Sample
--- | ---
Progress bar with percentage of the total transfer | `[########################::::::]  80%`
Bytes transferred | `19.17G`
Transfer speed | `86.65MB/s`
Elapsed time since starting rsync | `0:03:18`
Estimated Time to Completion | `ETA: 0:00:50`
Number of files transferred | `#306`
Files to scan/check<br>- percentage completed<br>- (number of files)<br>- spinner | `scan 46% (2410)\`

The spinner indicates that rsync is still checking if files need to be updated. Until this process completes the progress bar may decrease as new files are found.

## Limitations

Interactive ssh questions (host key, password) are happening outside of the piped content. That means that rsyncy does not know that ssh is waiting for input and will draw the status bar over it. If you have an idea on how to handle this case please let me know.

Workaround: connect once to your server via ssh to add it to the known_hosts file.

### ETA

The estimated time to completion is based on the percentage completion output by rsync, which is itself dependent on how many files rsync has scanned/found at a given point in time.

## Installation

```
pip3 install --user rsyncy

# or if you have pipx
pipx install rsyncy
```

On macOS you also need to `brew install rsync` because it ships with an rsync from 2006.

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

At the moment `rsyncy` itself has no options and only supports my preferred way of viewing rsync progress.

## Development

First record an rsync transfer with [pipevcr](https://github.com/laktak/pipevcr), then replay it to rsyncy when debugging.
