# Replay422ToPngConverter

[![Go Report Card](https://goreportcard.com/badge/github.com/Bendodroid/replay422ToPngConverter)](https://goreportcard.com/report/github.com/Bendodroid/replay422ToPngConverter)

## Installation

```shell script
go get github.com/Bendodroid/replay422ToPngConverter
```

Make sure that `$(go env GOPATH)/bin` is available in your `$PATH`.

## Usage

```text
$ replay422ToPngConverter -h
-h
    Display Help text
-j int
    Number of jobs to use for converting (default <nproc+2>)
-replayDir string
    A dir containing a replay.json and images (default ".")
-outputDir string
    Where to put the results (default ".")
```

Run the tool from anywhere and give paths as arguments or run it in the directory containing the replay.json directly.

## License

BSD 3-clause
