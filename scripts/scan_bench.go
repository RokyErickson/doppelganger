package main

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/pkg/errors"

	"github.com/spf13/pflag"

	"github.com/golang/protobuf/proto"

	"github.com/RokyErickson/doppelganger/cmd"
	"github.com/RokyErickson/doppelganger/cmd/profile"
	"github.com/RokyErickson/doppelganger/pkg/sync"
)

const (
	snapshotFile = "snapshot_test"
	cacheFile    = "cache_test"
)

var usage = `scan_bench [-h|--help] [-p|--profile] [-i|--ignore=<pattern>] <path>
`

func main() {
	flagSet := pflag.NewFlagSet("scan_bench", pflag.ContinueOnError)
	flagSet.SetOutput(ioutil.Discard)
	var ignores []string
	var enableProfile bool
	flagSet.StringSliceVarP(&ignores, "ignore", "i", nil, "specify ignore paths")
	flagSet.BoolVarP(&enableProfile, "profile", "p", false, "enable profiling")
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		if err == pflag.ErrHelp {
			fmt.Fprint(os.Stdout, usage)
			return
		} else {
			cmd.Fatal(errors.Wrap(err, "unable to parse command line"))
		}
	}
	arguments := flagSet.Args()
	if len(arguments) != 1 {
		cmd.Fatal(errors.New("invalid number of paths specified"))
	}
	path := arguments[0]

	fmt.Println("Analyzing", path)

	var profiler *profile.Profile
	var err error
	if enableProfile {
		if profiler, err = profile.New("scan_cold"); err != nil {
			cmd.Fatal(errors.Wrap(err, "unable to create profiler"))
		}
	}
	start := time.Now()
	snapshot, preservesExecutability, recomposeUnicode, cache, ignoreCache, err := sync.Scan(
		path, sha1.New(), nil, ignores, nil, sync.SymlinkMode_SymlinkPortable,
	)
	if err != nil {
		cmd.Fatal(errors.Wrap(err, "unable to create snapshot"))
	} else if snapshot == nil {
		cmd.Fatal(errors.New("target doesn't exist"))
	}
	stop := time.Now()
	if enableProfile {
		if err = profiler.Finalize(); err != nil {
			cmd.Fatal(errors.Wrap(err, "unable to finalize profiler"))
		}
		profiler = nil
	}
	fmt.Println("Cold scan took", stop.Sub(start))
	fmt.Println("Root preserves executability:", preservesExecutability)
	fmt.Println("Root requires Unicode recomposition:", recomposeUnicode)

	if enableProfile {
		if profiler, err = profile.New("scan_warm"); err != nil {
			cmd.Fatal(errors.Wrap(err, "unable to create profiler"))
		}
	}
	start = time.Now()
	snapshot, preservesExecutability, recomposeUnicode, _, _, err = sync.Scan(
		path, sha1.New(), cache, ignores, ignoreCache, sync.SymlinkMode_SymlinkPortable,
	)
	if err != nil {
		cmd.Fatal(errors.Wrap(err, "unable to create snapshot"))
	} else if snapshot == nil {
		cmd.Fatal(errors.New("target has been deleted since original snapshot"))
	}
	stop = time.Now()
	if enableProfile {
		if err = profiler.Finalize(); err != nil {
			cmd.Fatal(errors.Wrap(err, "unable to finalize profiler"))
		}
		profiler = nil
	}
	fmt.Println("Warm scan took", stop.Sub(start))
	fmt.Println("Root preserves executability:", preservesExecutability)
	fmt.Println("Root requires Unicode recomposition:", recomposeUnicode)

	if enableProfile {
		if profiler, err = profile.New("serialize_snapshot"); err != nil {
			cmd.Fatal(errors.Wrap(err, "unable to create profiler"))
		}
	}
	start = time.Now()
	buffer := proto.NewBuffer(nil)
	buffer.SetDeterministic(true)
	if err := buffer.Marshal(snapshot); err != nil {
		cmd.Fatal(errors.Wrap(err, "unable to serialize snapshot"))
	}
	serializedSnapshot := buffer.Bytes()
	stop = time.Now()
	if enableProfile {
		if err = profiler.Finalize(); err != nil {
			cmd.Fatal(errors.Wrap(err, "unable to finalize profiler"))
		}
		profiler = nil
	}
	fmt.Println("Snapshot serialization took", stop.Sub(start))

	if enableProfile {
		if profiler, err = profile.New("deserialize_snapshot"); err != nil {
			cmd.Fatal(errors.Wrap(err, "unable to create profiler"))
		}
	}
	start = time.Now()
	deserializedSnapshot := &sync.Entry{}
	if err = proto.Unmarshal(serializedSnapshot, deserializedSnapshot); err != nil {
		cmd.Fatal(errors.Wrap(err, "unable to deserialize snapshot"))
	}
	stop = time.Now()
	if enableProfile {
		if err = profiler.Finalize(); err != nil {
			cmd.Fatal(errors.Wrap(err, "unable to finalize profiler"))
		}
		profiler = nil
	}
	fmt.Println("Snapshot deserialization took", stop.Sub(start))

	start = time.Now()
	if err = deserializedSnapshot.EnsureValid(); err != nil {
		cmd.Fatal(errors.Wrap(err, "deserialized snapshot invalid"))
	}
	stop = time.Now()
	fmt.Println("Snapshot validation took", stop.Sub(start))

	start = time.Now()
	if err = ioutil.WriteFile(snapshotFile, serializedSnapshot, 0600); err != nil {
		cmd.Fatal(errors.Wrap(err, "unable to write snapshot"))
	}
	stop = time.Now()
	fmt.Println("Snapshot write took", stop.Sub(start))

	start = time.Now()
	if _, err = ioutil.ReadFile(snapshotFile); err != nil {
		cmd.Fatal(errors.Wrap(err, "unable to read snapshot"))
	}
	stop = time.Now()
	fmt.Println("Snapshot read took", stop.Sub(start))

	if err = os.Remove(snapshotFile); err != nil {
		cmd.Fatal(errors.Wrap(err, "unable to remove snapshot"))
	}

	fmt.Println("Serialized snapshot size is", len(serializedSnapshot), "bytes")

	fmt.Println(
		"Original/deserialized snapshots equivalent?",
		deserializedSnapshot.Equal(snapshot),
	)

	start = time.Now()
	sha1.Sum(serializedSnapshot)
	stop = time.Now()
	fmt.Println("SHA-1 snapshot digest took", stop.Sub(start))

	if enableProfile {
		if profiler, err = profile.New("serialize_cache"); err != nil {
			cmd.Fatal(errors.Wrap(err, "unable to create profiler"))
		}
	}
	start = time.Now()
	serializedCache, err := proto.Marshal(cache)
	if err != nil {
		cmd.Fatal(errors.Wrap(err, "unable to serialize cache"))
	}
	stop = time.Now()
	if enableProfile {
		if err = profiler.Finalize(); err != nil {
			cmd.Fatal(errors.Wrap(err, "unable to finalize profiler"))
		}
		profiler = nil
	}
	fmt.Println("Cache serialization took", stop.Sub(start))

	if enableProfile {
		if profiler, err = profile.New("deserialize_cache"); err != nil {
			cmd.Fatal(errors.Wrap(err, "unable to create profiler"))
		}
	}
	start = time.Now()
	deserializedCache := &sync.Cache{}
	if err = proto.Unmarshal(serializedCache, deserializedCache); err != nil {
		cmd.Fatal(errors.Wrap(err, "unable to deserialize cache"))
	}
	stop = time.Now()
	if enableProfile {
		if err = profiler.Finalize(); err != nil {
			cmd.Fatal(errors.Wrap(err, "unable to finalize profiler"))
		}
		profiler = nil
	}
	fmt.Println("Cache deserialization took", stop.Sub(start))

	start = time.Now()
	if err = ioutil.WriteFile(cacheFile, serializedCache, 0600); err != nil {
		cmd.Fatal(errors.Wrap(err, "unable to write cache"))
	}
	stop = time.Now()
	fmt.Println("Cache write took", stop.Sub(start))

	start = time.Now()
	if _, err = ioutil.ReadFile(cacheFile); err != nil {
		cmd.Fatal(errors.Wrap(err, "unable to read cache"))
	}
	stop = time.Now()
	fmt.Println("Cache read took", stop.Sub(start))

	if err = os.Remove(cacheFile); err != nil {
		cmd.Fatal(errors.Wrap(err, "unable to remove cache"))
	}

	fmt.Println("Serialized cache size is", len(serializedCache), "bytes")

	start = time.Now()
	if _, err = cache.GenerateReverseLookupMap(); err != nil {
		cmd.Fatal(errors.Wrap(err, "unable to generate reverse lookup map"))
	}
	stop = time.Now()
	fmt.Println("Reverse lookup map generation took", stop.Sub(start))
}
