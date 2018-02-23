# go-sshd

A simple sshd implemented in golang, with [asciicast](https://asciinema.org/) support.

Currently it only implemented interactive shell through ssh. Functions like
running command, ssh forwarding are not will come later.

There is likely to be some security issue with this project at the moment. Make
sure not to run this on production server.

## Installation

```bash
go get github.com/inoc603/go-sshd/cmd
```

## Developement

Build the project with:

```
make build
```

Run the built program:

```
make fg
```

## Usage

> TODO

