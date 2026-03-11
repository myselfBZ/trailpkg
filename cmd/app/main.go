package main

import (
	"fmt"
	"os"

	"github.com/myselfBZ/trailpkg/internal/manifest"
)



func main() {
	
	userInput := newUserInput(os.Args)

	app := app{ input: userInput,
		manifestManager: manifest.NewManifestManager("/Users/bobirmirzo/trailpkg"),
	}

	exitStatus := app.executeUserInput()

	if exitStatus.Code != StatusSuccess {
		fmt.Println(exitStatus.Message)
		os.Exit(1)
	}
	
	// End
}
