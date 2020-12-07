package main

import (
	"os"
	"flag"
	"bufio"
	"fmt"
	"sort"

	"github.com/aewens/intcode-vm/pkg/shared"
	"github.com/aewens/intcode-vm/pkg/intcode"
)

type Modes struct {
	Verbose bool
}

func catch(err error) {
	if err != nil {
		panic(err)
	}
}

func readData() (string, *Modes) {
	fileFlag := flag.String("f", "-", "Path of program to run")
	verboseFlag := flag.Bool("v", false, "Use verbose mode")

	flag.Parse()

	var scanner *bufio.Scanner
	if *fileFlag == "-" {
		scanner = bufio.NewScanner(os.Stdin)
		scanner.Scan()
	} else {
		file, err := os.Open(*fileFlag)
		catch(err)

		defer file.Close()
		scanner = bufio.NewScanner(file)
		scanner.Scan()
	}

	program := scanner.Text()

	err := scanner.Err()
	catch(err)

	return program, &Modes{
		Verbose: *verboseFlag,
	}
}

func main() {
	defer shared.Cleanup()
	shared.HandleSigterm()

	program, modes := readData()
	computer := intcode.New(program)

	if modes.Verbose {
		fmt.Println("Output:")
	}
	result := computer.Run()

	if modes.Verbose {
		addresses := []int{}
		for address := range result {
			addresses = append(addresses, address)
		}
		sort.Ints(addresses)

		fmt.Println("-------\nMemory:")
		for key := range addresses {
			value := result[key]
			fmt.Printf("%d:\t%d\n", key, value)
		}
	}
}
