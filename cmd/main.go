package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"strings"
)

const (
	KiB = 1024
)

var (
	humanReadable = flag.Bool("h", false, "Output sizes in MiB, GiB...")
	printTotal    = flag.Bool("t", false, "Output a total line")
	level         = flag.Int("l", -1, "Define till which level the output should be printed")
)

func prettifyOutput(size float64, suffix string) string {
	return fmt.Sprintf("%.2f %s", size/KiB, suffix)
}

func prettyByteSize(b int64) string {
	bf := float64(b)
	for _, unit := range []string{"B  ", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB"} {
		if math.Abs(bf) < float64(KiB) {
			return fmt.Sprintf("%.2f %s", bf, unit)
		}
		bf /= float64(KiB)
	}
	return fmt.Sprintf("%.2f YiB", bf)
}

func iterDirs(entries []os.DirEntry, path string, level int, humanReadable bool) int64 {
	var totalSize int64
	subLvl := level

	if subLvl > 0 {
		subLvl--
	}
	for _, entry := range entries {
		if entry.IsDir() {
			i, err := entry.Info()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			dirSize := i.Size()

			subPath := strings.ReplaceAll(fmt.Sprintf("%s/%s", path, entry.Name()), "//", "/")
			subEntries, err := os.ReadDir(subPath)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}

			dirSize += iterDirs(subEntries, subPath, subLvl, humanReadable)
			if subLvl > 0 || subLvl == -1 {
				dirSizeStr := prettifyOutput(float64(dirSize), "B  ")
				if humanReadable {
					dirSizeStr = prettyByteSize(dirSize)
				}
				fmt.Printf("%s\t%s\n", dirSizeStr, subPath)
			}
			totalSize += dirSize
		} else {
			info, err := entry.Info()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			if info.Name() == "kcore" && info.Size() > 10485760 {
				continue
			}
			size := info.Size()
			totalSize += size
		}
	}
	return totalSize
}

func main() {
	flag.Parse()
	path := "."
	if len(flag.Args()) > 0 {
		path = flag.Arg(0)
		if len(path) > 1 {
			path = strings.TrimRight(path, "/")
		}
	}

	dirs, err := os.ReadDir(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	totalSize := iterDirs(dirs, path, *level+1, *humanReadable)
	var total string
	if *humanReadable {
		total = prettyByteSize(totalSize)
	} else {
		total = prettifyOutput(float64(totalSize), "B  ")
	}
	fmt.Printf("%s\t%s\n", total, path)
	if *printTotal {
		fmt.Printf("%s\tTotal\n", total)
	}

}
