package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/samber/lo"
)

func main() {
	in := flag.String("in", "", "Input file or directory")
	out := flag.String("out", "", "Output file or directory")
	luts := flag.String("luts", "", "LUTs to apply")
	eq := flag.String("eq", "", "Equalizer to apply, example: contrast=1:brightness=0:saturation=1")
	metadata := flag.Bool("metadata", true, "Copy metadata")
	sips := flag.Bool("sips", true, "Convert JPG to heic by sips")
	imagemagick := flag.Bool("imagemagick", false, "Convert JPG to heic by imagemagick")

	flag.Parse()

	if *in == "" || *out == "" {
		flag.Usage()
		return
	}

	if *in == *out {
		fmt.Println("input and output cannot be the same")
		return
	}

	lo.Must0(os.MkdirAll(*out, os.ModePerm))

	lo.Must0(filepath.Walk(*in, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		outFile := filepath.Join(*out, filepath.Base(path))

		var filter []string
		if *luts == "" {
			filter = append(filter, fmt.Sprintf("lut3d=%s", os.TempDir()+"Neutral A7s3 Sl2sg3c.cube"))
		} else {
			for _, lut := range strings.Split(*luts, ",") {
				filter = append(filter, fmt.Sprintf("lut3d=%s", lut))
			}
		}

		if *eq != "" {
			filter = append(filter, fmt.Sprintf("eq=%s", *eq))
		}

		cmds := []string{
			fmt.Sprintf("ffmpeg -i %s -filter_complex \"%s\" -q:v 1 -qmin 1 -qmax 1 -y %s -hide_banner -loglevel quiet",
				path, strings.Join(filter, ","), outFile),
		}

		if *metadata {
			cmds = append(cmds, fmt.Sprintf("exiftool -overwrite_original -tagsfromfile %s -all:all %s", path, outFile))
		}

		if *sips {
			cmds = append(cmds, fmt.Sprintf("sips -s format heic %s --out %s.heic", outFile, outFile))
			cmds = append(cmds, fmt.Sprintf("rm %s", outFile))
		}

		for _, cmd := range cmds {
			lo.Must0(runCmd(cmd), cmd)
		}

		return nil
	}))

	if *imagemagick {
		lo.Must0(runCmd(fmt.Sprintf("cd %s && magick mogrify -format heic -depth 10 *.JPG", *out)))
	}

	fmt.Println("finished!")
}

func runCmd(cmd string) error {
	fmt.Println(cmd)

	do := exec.Command("bash", "-c", cmd)
	do.Stderr = os.Stderr
	do.Stdout = os.Stdout
	err := do.Run()
	if err != nil {
		fmt.Println("process failed", err)
		return err
	}

	return nil
}
