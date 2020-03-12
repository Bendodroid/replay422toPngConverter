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

	"github.com/Bendodroid/replay422toPngConverter/logic"
	"github.com/Bendodroid/replay422toPngConverter/models"
)

// CLI flags
var cpuProfile, memProfile, outputDir, replayDir string
var nJobs, pngCompression int
var modifyReplayJson, help bool

func init() {
	flag.StringVar(&cpuProfile, "cpuprofile", "", "Write cpu profiler data to `file`")
	flag.StringVar(&memProfile, "memprofile", "", "Write memory profiler data to `file`")
	flag.StringVar(&outputDir, "outputDir", ".", "Where to put the results")
	flag.StringVar(&replayDir, "replayDir", ".", "A dir containing a replay.json and images")
	flag.BoolVar(&modifyReplayJson, "i", false, "Whether to modify the original replay.json")
	flag.BoolVar(&help, "h", false, "Display Help text")
	flag.IntVar(&nJobs, "j", runtime.NumCPU()+2, "Number of jobs to use for converting (max: 255)")
	flag.IntVar(&pngCompression, "c", -1, "See https://godoc.org/image/png#CompressionLevel")
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

	replayDirAbs := filepath.Clean(replayDir)
	fmt.Println(replayDirAbs)
	superJob := models.SuperJob{Dir: replayDirAbs}

	logic.BuildSuperJob(&superJob, nJobs, pngCompression, modifyReplayJson)

	logic.HandleSuperJob(&superJob)

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
