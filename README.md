
# rsyncy

A status bar for [rsync](https://github.com/WayneD/rsync).

I love rsync but I always felt it was either too chatty when transferring lots of small files or did not show enough information for the large files in between.

rsyncy is a fun experiment to fix this without having to bother the rsync developers.

![rsyncy](readme/demo.gif)

## Installation

```
pip3 install --user rsyncy

# or if you have pipx
pipx install rsyncy
```

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
$ rsyncy -a --info=progress2 -hv FROM/ TO | rsyncy
```

At the moment `rsyncy` itself has no options and only supports my preferred way of viewing rsync progress.

## Development

First record an rsync transfer with [pipevcr](https://github.com/laktak/pipevcr), then replay it to rsyncy when debugging.
