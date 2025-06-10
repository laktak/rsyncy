
# rsyncy (python version)

A status/progress bar for [rsync](https://github.com/WayneD/rsync).

![gif of rsyncy -a a/ b](https://raw.githubusercontent.com/laktak/rsyncy/readme/readme/demo-y.gif "rsyncy -a a/ b")


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

See https://github.com/laktak/rsyncy


## Usage

`rsyncy` is a wrapper around `rsync`.

- You run `rsyncy` with the same arguments as it will pass them to `rsync` internally.
- Do not specify any `--info` arguments, rsyncy will automatically add `--info=progress2` and `-hv` internally.

```
# simple example
$ rsyncy -a FROM/ TO
```

