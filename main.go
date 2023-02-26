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
	luts := flag.String("luts", "HLG3 for Rec709.cube", "LUTs to apply")
	metadata := flag.Bool("metadata", true, "Copy metadata")

	flag.Parse()

	if *in == "" || *out == "" {
		flag.Usage()
		return
	}

	if *in == *out {
		fmt.Println("input and output cannot be the same")
		return
	}

	if *luts == "" {
		fmt.Println("luts cannot be empty")
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
		for _, lut := range strings.Split(*luts, ",") {
			filter = append(filter, fmt.Sprintf("lut3d=%s", os.TempDir()+lut))
		}

		cmds := []string{
			fmt.Sprintf("ffmpeg -i %s -filter_complex \"%s\" -q:v 1 -qmin 1 -qmax 1 -y %s -hide_banner -loglevel quiet",
				path, strings.Join(filter, ","), outFile),
		}

		if *metadata {
			cmds = append(cmds, fmt.Sprintf("exiftool -overwrite_original -tagsfromfile %s -all:all %s", path, outFile))
		}

		for _, cmd := range cmds {
			lo.Must0(runCmd(cmd), cmd)
		}

		return nil
	}))

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
