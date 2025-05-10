// Generated wrapper for executing Add
package main

import (
	"encoding/json"
	"fmt"
	"os"

	// Import the package containing the function
	pkg "github.com/test/simplemath"
)

func main() {
	// Call the function

	result := pkg.Add(5, 3)

	// Encode the result to JSON and print it
	jsonResult, err := json.Marshal(result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling result: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonResult))

}
