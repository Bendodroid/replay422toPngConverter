# Replay422ToPngConverter

[![Go Report Card](https://goreportcard.com/badge/github.com/Bendodroid/replay422ToPngConverter)](https://goreportcard.com/report/github.com/Bendodroid/replay422ToPngConverter)

## Installation

```shell script
go get github.com/Bendodroid/replay422ToPngConverter
```

Make sure that `$(go env GOPATH)/bin` is available in your `$PATH`.

## Usage

```text
$ replay422toPngConverter -h
  -h	Display Help text
  -i	Whether to modify the original replay.json
  -j int
    	Number of jobs to use for converting (default $(nproc+2))
  -c int
    	-c=X  See https://godoc.org/image/png#CompressionLevel (default -1)
  -inputPath string
    	Path to a replay.json or a folder containing data from multiple robots (default ".")
  -outputDir string
    	Where to put the results (default ".")
```

Run the tool from anywhere and give paths as arguments or run it from the directory containing the replay.json.  
`inputPath` can be one of the following:

- `something/something/replay.json` In this case, the tool converts data for this robot only.
- `something/something/replay_* (dir)` In this case, the tool appends `replay.json` to the path and falls back to the first case.
- `something/something/2020-23-42_ReplayData (dir)` If the dir contains at least one dir matching `10.1.24.*`, this is treated as data from multiple robots.
  The tool automatically finds a `replay.json` for each dataset and converts them sequentially.

When specifying outputDir in the third case, directory structure is preserved in the target directory.

## License

BSD 3-clause
