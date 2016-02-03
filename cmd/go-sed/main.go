package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

        "github.com/waywardcode/sed"
)

var noPrint bool
var evalProg string
var sedFile string

func init() {
	flag.BoolVar(&noPrint, "n", false, "do not automatically print lines")
	flag.BoolVar(&noPrint, "silent", false, "do not automatically print lines")
	flag.BoolVar(&noPrint, "quiet", false, "do not automatically print lines")

	flag.StringVar(&evalProg, "e", "", "a string to evaluate as the program")
	flag.StringVar(&evalProg, "expression", "", "a string to evaluate as the program")

	flag.StringVar(&sedFile, "f", "", "a file to read as the program")
	flag.StringVar(&sedFile, "file", "", "a file to read as the program")
}

func compileScript(args *[]string) (*sed.Engine, error) {
	var program *bufio.Reader

	// STEP ONE: Find the script
	switch {
	case evalProg != "":
		program = bufio.NewReader(strings.NewReader(evalProg))
		if sedFile != "" {
			return nil, fmt.Errorf("Cannot specify both an expression and a program file!")
		}
	case sedFile != "":
		fl, err := os.Open(sedFile)
		if err != nil {
			return nil, fmt.Errorf("Error opening %s: %v", sedFile, err)
		}
		defer fl.Close()
		program = bufio.NewReader(fl)
	case len(*args) > 0:
		// no -e or -f given, so the first argument is taken as the script to run
		program = bufio.NewReader(strings.NewReader((*args)[0]))
		*args = (*args)[1:]
	}

	// STEP TWO: compile the program
        var compiler func(*bufio.Reader) (*sed.Engine,error) 
	if(noPrint) {
		compiler = sed.NewQuiet
        } else {
		compiler = sed.New
        }
	return compiler(program)
}

func main() {
	flag.Parse()
	args := flag.Args()
	var err error

	// Find and compile the script
	engine, err := compileScript(&args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}

	// Now, run the script against the input
	output := bufio.NewWriter(os.Stdout)

	if len(args) == 0 {
		err = engine.Run(bufio.NewReader(os.Stdin), output)
	} else {
		for _, fname := range args {
			fl, err := os.Open(fname)
			if err != nil {
				break
			}

			err = engine.Run(bufio.NewReader(fl), output)

			fl.Close()
			if err != nil {
				break
			}
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}