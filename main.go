package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

const Version = "v0.0.1"

func main() {
	var version bool

	root := &cobra.Command{
		Use:          "gs-gen",
		Short:        "gen go server code from idl files",
		SilenceUsage: true,
	}

	root.Flags().BoolVar(&version, "version", false, "show version")

	root.RunE = func(cmd *cobra.Command, args []string) error {
		if version {
			fmt.Println(root.Short)
			fmt.Println(Version)
			return nil
		}

		if _, err := os.Stat("gs.json"); err != nil {
			log.Fatalln("gs.json not found")
		}

		entries, err := os.ReadDir("idl")
		if err != nil {
			log.Fatalln(err)
		}

		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			switch e.Name() {
			case "http":
				genHttp()
			default: // for linter
			}
		}

		return nil
	}

	if err := root.Execute(); err != nil {
		os.Exit(-1)
	}
}

func genHttp() {
	currDir, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	dir := filepath.Join(currDir, "idl/http/proto")
	if err = os.RemoveAll(dir); err != nil {
		log.Fatalln(err)
	}
	if err = os.MkdirAll(dir, os.ModePerm); err != nil {
		log.Fatalln(err)
	}

	r, w, err := os.Pipe()
	if err != nil {
		log.Fatalln(err)
	}

	go func() {
		f := bufio.NewReader(r)
		for {
			line, _, err := f.ReadLine()
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Print(string(line))
		}
		r.Close()
	}()

	cmd := exec.Command("gs-http-gen", "--server", "--output", dir)
	cmd.Dir = filepath.Dir(dir)
	cmd.Stdout = w
	cmd.Stderr = w
	if err = cmd.Run(); err != nil {
		log.Fatalln(err)
	}
	w.Close()
}
