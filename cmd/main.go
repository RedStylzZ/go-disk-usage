package main

import (
	"fmt"
	"log"
	"math"
	"os"
)

const (
	KiB = 1024
	MiB = KiB * 1024
	GiB = MiB * 1024
	TiB = GiB * 1024
	PiB = TiB * 1024
	KB  = 1000
	MB  = KB * 1000
	GB  = MB * 1000
	TB  = GB * 1000
	PB  = TB * 1000
)

func prettifyOutput(size float64, suffix string) string {
	return fmt.Sprintf("%.2f %s", float64(size)/KiB, suffix)
}

func prettyByteSize(b int64) string {
	bf := float64(b)
	for _, unit := range []string{"B  ", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB"} {
		if math.Abs(bf) < 1024.0 {
			return fmt.Sprintf("%.2f %s", bf, unit)
		}
		bf /= 1024.0
	}
	return fmt.Sprintf("%.2f YiB", bf)
}

func iterDirs(entries []os.DirEntry, path string, level int, humanReadable bool) int64 {
	var totalSize int64
	level--
	for _, entry := range entries {
		var dirSize int64
		var info os.FileInfo
		var err error
		if entry.IsDir() {
			subPath := fmt.Sprintf("%s/%s/", path, entry.Name())
			subEntries, err := os.ReadDir(subPath)
			if err != nil {
				fmt.Println(err)
				continue
			}
			dirSize += iterDirs(subEntries, subPath, level, humanReadable)
			if level >= 0 {
				dirSizeStr := prettifyOutput(float64(dirSize)/KiB, "KiB")
				if humanReadable {
					dirSizeStr = prettyByteSize(dirSize)
				}
				fmt.Printf("%s\t%s\n", dirSizeStr, entry.Name())
			}
			totalSize += dirSize
		} else {
			info, err = entry.Info()
			if err != nil {
				log.Fatal(err)
			}
			totalSize += info.Size()
		}
	}
	return totalSize
}

func main() {
	level := 1
	humanReadable := true
	path := ""
	dirs, _ := os.ReadDir(path)
	totalSize := iterDirs(dirs, path, level, humanReadable)
	if humanReadable {
		fmt.Printf("%s\tTotal\n", prettyByteSize(totalSize))
	} else {
		output := prettifyOutput(float64(totalSize)/KiB, "KiB")
		fmt.Printf("%s\tTotal", output)
	}

}
