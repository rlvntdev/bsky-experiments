package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ericvolp12/bsky-experiments/pkg/graph"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run main.go inputfile outputfile")
		return
	}

	inputFile := os.Args[1]
	outputFile := os.Args[2]

	binReaderWriter := graph.BinaryGraphReaderWriter{}

	// Read the graph from the Binary file
	g, err := binReaderWriter.ReadGraph(inputFile)
	if err != nil {
		log.Fatalf("Error reading graph from binary file: %v", err)
	}

	// Write the graph to the new Binary database
	err = binReaderWriter.WriteGraph(g, outputFile)
	if err != nil {
		log.Fatalf("Error writing graph to binary file: %v", err)
	}

	fmt.Println("Graph successfully written to binary file")
}