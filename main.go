package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/Bendodroid/replay422toPngConverter/models"
	"github.com/Bendodroid/replay422toPngConverter/util"
)

// CLI flags
var cpuProfile, memProfile, inputPath, outputDir string
var nJobs, pngCompression int
var modifyReplayJson, help bool

func init() {
	flag.StringVar(&cpuProfile, "cpuProfile", "", "Write cpu profiler data to `file`")
	flag.StringVar(&memProfile, "memProfile", "", "Write memory profiler data to `file`")
	flag.StringVar(&inputPath, "inputPath", ".", "Path to a replay.json or a folder containing data from multiple robots")
	flag.StringVar(&outputDir, "outputDir", ".", "Where to put the results")
	flag.IntVar(&nJobs, "j", runtime.NumCPU()+4, "Number of jobs to use for converting")
	flag.IntVar(&pngCompression, "c", -1, "-c=X  See https://godoc.org/image/png#CompressionLevel")
	flag.BoolVar(&modifyReplayJson, "i", false, "Whether to modify the original replay.json")
	flag.BoolVar(&help, "h", false, "Display Help text")
	flag.Parse()
	// If you just want the manual
	if help {
		flag.PrintDefaults()
		os.Exit(0)
	}
}

func main() {
	startTime := time.Now()

	if cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	inputPath = util.ExpandPath(inputPath)
	outputDir = util.ExpandPath(outputDir)
	log.Printf("Parsed %s as input and %s as output...", inputPath, outputDir)

	if fileInfo, err := os.Stat(inputPath); err == nil {
		if !fileInfo.IsDir() && filepath.Base(inputPath) == "replay.json" {
			// Path is something/something/replay.json
			log.Printf("%s seems to be a path to a single replay.json. Using %s as output dir.", inputPath, outputDir)
			var rj models.RobotJob
			rj.PrePrepare(inputPath, outputDir, nJobs, pngCompression, modifyReplayJson)
			rj.Prepare()
			rj.Run()
			goto Exit
		} else if fileInfo.IsDir() {
			// It's a dir
			if x, _ := filepath.Match("replay_*", fileInfo.Name()); x {
				// It's a replay_* dir
				log.Printf("%s seems to be a path to a replay_* dir. Using %s as output dir.", inputPath, outputDir)
				var rj models.RobotJob
				inputPath = filepath.Join(inputPath, "replay.json")
				rj.PrePrepare(inputPath, outputDir, nJobs, pngCompression, modifyReplayJson)
				rj.Prepare()
				rj.Run()
				goto Exit
			} else if f := util.MatchPatternInPath(inputPath, "10.1.24.*"); f != nil {
				// It's a dir for multiple robots
				log.Printf("%s seems to be a path to data from multiple robots. Using %s as output dir base.", inputPath, outputDir)
				var superJob models.CollectionJob
				superJob.PrePrepare(inputPath, outputDir, nJobs, pngCompression, modifyReplayJson)
				superJob.Prepare()
				superJob.Run()
				goto Exit
			}
			log.Fatalf("%s is a dir but I can't make sense of it. Find someone responsible.", inputPath)
		}
	} else if os.IsNotExist(err) {
		log.Fatalf("File path %s does not exist!", inputPath)
	} else {
		log.Fatalf("Error occured when trying to open %s: \n%s", inputPath, err)
	}

Exit:
	// Finish and print time it took
	fmt.Println("Finished converting after ", time.Since(startTime).String(), "!")

	if memProfile != "" {
		f, err := os.Create(memProfile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}
