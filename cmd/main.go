package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

const (
	KiB float64 = 1024
	MiB         = KiB * 1024
	GiB         = MiB * 1024
	TiB         = GiB * 1024
	PiB         = TiB * 1024
)

type Arguments struct {
	HumanReadable bool
	PrintTotal    bool
	depth         int
	ShowFiles     bool
	Threshold     float64
	Endline       string
	PrintErrs     bool
}

var (
	args          Arguments
	humanReadable = flag.Bool("h", false, "Output sizes in MiB, GiB...")
	printTotal    = flag.Bool("t", false, "Output a total line")
	depth         = flag.Int("l", -1, "Define till which depth the output should be printed")
	showFiles     = flag.Bool("f", false, "Use if you want to output the files")
	threshold     = flag.String("th", "0K", "Define a threshold for the minimum size to be printed")
	nulEnd        = flag.Bool("0", false, "Print the results with an NUL character insted of newline")
	printErrs     = flag.Bool("ne", false, "Use to surpress the output of errors to the cli output")
)

func prettifyOutput(size float64, suffix string) string {
	return fmt.Sprintf("%.2f %s", size, suffix)
}

func getAsKibibyte(size float64) string {
	return prettifyOutput(size/KiB, "KiB")
}

func prettyByteSize(b float64) string {
	for _, unit := range []string{"B  ", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB"} {
		if math.Abs(b) < KiB {
			return prettifyOutput(b, unit)
		}
		b /= KiB
	}
	return prettifyOutput(b, "YiB")
}

func getSizeStr(size float64) string {
	sizeStr := getAsKibibyte(size)
	if args.HumanReadable {
		sizeStr = prettyByteSize(size)
	}
	return sizeStr
}

func iterDirs(entries []os.DirEntry, path string, depth int) float64 {
	var totalSize float64

	if depth > 0 {
		depth--
	}
	for _, entry := range entries {
		var isPrintAllowed bool

		info, err := entry.Info()
		if err != nil {
			if args.PrintErrs {
				fmt.Fprintln(os.Stderr, err)
			}
			continue
		}

		path := strings.ReplaceAll(fmt.Sprintf("%s/%s", path, entry.Name()), "//", "/")

		size := float64(info.Size())
		if entry.IsDir() {
			subEntries, err := os.ReadDir(path)
			if err != nil {
				if args.PrintErrs {
					fmt.Fprintln(os.Stderr, err)
				}
				continue
			}

			size += iterDirs(subEntries, path, depth)
			if depth > 0 || depth == -1 {
				isPrintAllowed = true
			}
		} else {
			if info.Name() == "kcore" && info.Size() > 10485760 {
				continue
			}
			if args.ShowFiles && (depth > 0 || depth == -1) {
				isPrintAllowed = true
			}
		}

		totalSize += size
		if isPrintAllowed && size > args.Threshold {
			fmt.Printf("%s\t%s%s", getSizeStr(size), path, args.Endline)
		}
	}
	return totalSize
}

func main() {
	flag.Parse()

	var thSize float64
	thStr := *threshold
	thSuffix := thStr[len(thStr)-1]
	th, err := strconv.Atoi(thStr[:len(thStr)-1])
	if err != nil {
		if *printErrs {
			fmt.Fprintf(os.Stderr, "invalid threshold %s: %v\n", thStr, err)
		}
		return
	}

	floatTh := float64(th)
	switch thSuffix {
	case 'B':
		thSize = floatTh
	case 'K':
		thSize = floatTh * KiB
	case 'M':
		thSize = floatTh * MiB
	case 'G':
		thSize = floatTh * GiB
	case 'T':
		thSize = floatTh * TiB
	case 'P':
		thSize = floatTh * PiB
	default:
		fmt.Fprintf(os.Stderr, "invalid threshold %s: %v\n", thStr, err)
		return
	}

	endline := "\n"
	if *nulEnd {
		endline = "\x00"
	}

	args = Arguments{
		HumanReadable: *humanReadable,
		depth:         *depth,
		PrintTotal:    *printTotal,
		ShowFiles:     *showFiles,
		Threshold:     thSize,
		Endline:       endline,
	}

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

	if *depth > -1 {
		*depth++
	}
	totalSize := float64(iterDirs(dirs, path, *depth))

	var total string
	if *humanReadable {
		total = prettyByteSize(totalSize)
	} else {
		total = getAsKibibyte(float64(totalSize))
	}

	fmt.Printf("%s\t%s%s", total, path, args.Endline)
	if *printTotal {
		fmt.Printf("%s\tTotal%s", total, args.Endline)
	}

}
