// Copyright (c) 2023-2024, Nubificus LTD
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/base64"
	"encoding/json"
	"context"
	"flag"
	"os"
	"fmt"
	"bytes"
	"strings"
	"io/ioutil"

	"github.com/moby/buildkit/client/llb"
	ocispecs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

const (
	currentWD          string = "client-WD"
	unikraftKernelPath string = "/unikraft/bin/kernel"
	unikraftHub        string = "unikraft.org"
)

type CLIOpts struct {
	ContainerFile    string // The Containerfile to be used for building the unikernel container
}

type PackInstructions struct {
	Base   string			  // The Base image to use
	Copies []instructions.CopyCommand // Copy commands
	Annots map[string]string	  // Annotations
}

func usage() {

	fmt.Println("Usage of bun")
	fmt.Printf("%s [<args>]\n\n", os.Args[0])
	fmt.Println("Supported command line arguments")
	fmt.Println("\t-f, --file filename \t\tPath to the Containerfile")
}

func parseCLIOpts() CLIOpts {
	var opts CLIOpts

	flag.StringVar(&opts.ContainerFile, "file", "", "Path to the Containerfile")
	flag.StringVar(&opts.ContainerFile, "f", "", "Path to the Containerfile")

	flag.Usage = usage
	flag.Parse()

	return opts
}

func parseFile(file string) (*PackInstructions, error) {
	var instr *PackInstructions
	instr = new(PackInstructions)
	instr.Annots = make(map[string]string)

	CntrFileContent, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("Failed to read %s: %v\n", file, err)
		return nil, err
	}

	// Parse the Dockerfile
	r := bytes.NewReader(CntrFileContent)
	parseRes, err := parser.Parse(r)
	if err != nil {
		fmt.Printf("Failed to parse  %s: %v\n", file, err)
		return nil, err
	}

	// Traverse Dockerfile commands
	for _, child := range parseRes.AST.Children {
		cmd, err := instructions.ParseInstruction(child)
		if err != nil {
			fmt.Printf("Failed to parse instruction %s: %v\n", child.Value, err)
			return nil, err
		}
		switch c := cmd.(type) {
		case *instructions.Stage:
			// Handle FROM
			if instr.Base != "" {
				return nil, fmt.Errorf("Multi-stage builds are not supported")
			}
			instr.Base = c.BaseName
		case *instructions.CopyCommand:
			// Handle COPY
			instr.Copies = append(instr.Copies, *c)
		case *instructions.LabelCommand:
			// Handle LABLE annotations
			for _, kvp := range c.Labels {
				annotKey := strings.Trim(kvp.Key, "\"")
				instr.Annots[annotKey] = strings.Trim(kvp.Value, "\"")
			}
		case instructions.Command:
			// Catch all other commands
			fmt.Printf("UNsupported command%s\n", c.Name())
		default:
			fmt.Printf("%f is not a command type\n", c)
		}

	}

	return instr, nil
}

func copyIn(base llb.State, from string, src string, dst string) llb.State {
	var copyState llb.State
	var localSrc llb.State

	localSrc = llb.Local("client-WD")
	copyState = base.File(llb.Copy(localSrc, src, dst, &llb.CopyInfo{
				CreateDestPath: true,}))

	return copyState
}

func main() {
	var cliOpts CLIOpts
	var base llb.State
	var outState llb.State
	var packInst *PackInstructions

	cliOpts = parseCLIOpts()

	if cliOpts.ContainerFile == "" {
		fmt.Println("Please specify the Containerfile")
		fmt.Println("Use -h or --help for more info")
		os.Exit(1)
	}

	packInst, err := parseFile(cliOpts.ContainerFile)
	if err != nil {
		fmt.Println("Error parsing packing instructions", err)
		os.Exit(1)
	}
	for annot, val := range packInst.Annots {
		encoded := base64.StdEncoding.EncodeToString([]byte(val))
		packInst.Annots[annot] = string(encoded)
	}
	byteObj, err := json.Marshal(packInst.Annots)
	if err != nil {
		fmt.Println("Failed to marshal urunc annotations: %v", err)
		os.Exit(1)
	}
	if packInst.Base == "scratch" {
		base = llb.Scratch()
	} else if strings.HasPrefix(packInst.Base, unikraftHub) {
		// Define the platform to qemu/amd64 so we cna pull unikraft images
		platform := ocispecs.Platform{
			OS:           "qemu",
			Architecture: "amd64",
		}
		base = llb.Image(packInst.Base, llb.Platform(platform),)
	} else {
		base = llb.Image(packInst.Base)
	}

	for _, aCopy := range packInst.Copies {
		base = copyIn(base, currentWD, aCopy.SourcePaths[0], aCopy.DestPath)
	}
	outState = base.File(llb.Mkfile("/urunc.json", 0644, byteObj))
	dt, err := outState.Marshal(context.TODO(), llb.LinuxAmd64)
	if err != nil {
		panic(err)
	}
	llb.WriteTo(dt, os.Stdout)
}
