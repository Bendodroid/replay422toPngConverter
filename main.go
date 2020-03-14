package main

import (
	"flag"
	"fmt"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/Bendodroid/replay422toPngConverter/logic"
	"github.com/Bendodroid/replay422toPngConverter/models"
	"github.com/Bendodroid/replay422toPngConverter/util"
)

// CLI flags
var cpuProfile, memProfile, outputDir, replayPath string
var nJobs, pngCompression int
var modifyReplayJson, help bool

func init() {
	// TODO Update README!!!
	flag.StringVar(&cpuProfile, "cpuprofile", "", "Write cpu profiler data to `file`")
	flag.StringVar(&memProfile, "memprofile", "", "Write memory profiler data to `file`")
	flag.StringVar(&outputDir, "outputDir", ".", "Where to put the results")
	flag.StringVar(&replayPath, "replayPath", ".", "A dir containing a replay.json and images") // TODO Adjust message
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

	replayPathAbs := util.ExpandHome(filepath.Clean(replayPath))
	log.Printf("Using %s as absolute path to replay Data", replayPathAbs)

	if fileInfo, err := os.Stat(replayPathAbs); err == nil {
		if !fileInfo.IsDir() && filepath.Base(replayPathAbs) == "replay.json" {
			// Path is ....../replay.json
			rjc := models.ReplayJsonContainer{
				ReplayDirAbs:     filepath.Dir(replayPathAbs),
				ReplayJsonPath:   replayPathAbs,
				OutputDir:        outputDir,
				CompressionLevel: png.CompressionLevel(pngCompression),
				NJobs:            nJobs,
				ModifyOriginal:   modifyReplayJson,
			}
			util.LoadReplayJson(replayPathAbs, &rjc.ReplayJson)
			logic.HandleReplayJSON(&rjc)
		} else if fileInfo.IsDir() {
			// It's a dir
			if x, _ := filepath.Match("replay_*", fileInfo.Name()); x {
				// It's a path to a replay_* dir
				replayPathAbs = filepath.Join(replayPathAbs, "replay.json")
				rjc := models.ReplayJsonContainer{
					ReplayDirAbs:     filepath.Dir(replayPathAbs),
					ReplayJsonPath:   replayPathAbs,
					OutputDir:        outputDir,
					CompressionLevel: png.CompressionLevel(pngCompression),
					NJobs:            nJobs,
					ModifyOriginal:   modifyReplayJson,
				}
				util.LoadReplayJson(replayPathAbs, &rjc.ReplayJson)
				logic.HandleReplayJSON(&rjc)
			} else if f, _ := util.MatchPatternInPath(replayPathAbs, "10.1.24.*"); f != nil {
				// It's for multiple robots
				superJob := models.SuperJob{Dir: replayPathAbs}
				logic.BuildSuperJob(&superJob, nJobs, pngCompression, modifyReplayJson)
				logic.HandleSuperJob(&superJob)
			}
		}
	} else if os.IsNotExist(err) {
		log.Fatalf("File path %s does not exist!", replayPathAbs)
	} else {
		log.Fatalf("Error occured when trying to open %s: \n%s", replayPathAbs, err)
	}

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
