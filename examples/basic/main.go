// Basic example demonstrating core LocalFS functionality
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ericsperano/localfs"
)

// Define file types
const (
	LeagueFileType localfs.FileType = iota
)

// LeagueFile implements the File interface
type LeagueFile struct {
	season int
}

func (f LeagueFile) Dir() string {
	return fmt.Sprintf("%d/league", f.season)
}

func (f LeagueFile) Name() string {
	return "league"
}

func (f LeagueFile) Ext() string {
	return "json"
}

func main() {
	// Create a temporary directory for this example
	tmpDir, err := os.MkdirTemp("", "localfs-example-*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	fmt.Printf("Using directory: %s\n\n", tmpDir)

	// Create LocalFS instance
	lfs := localfs.New(tmpDir)

	// Register the file type
	lfs.RegisterFileType(LeagueFileType, func(args ...any) localfs.File {
		return LeagueFile{season: args[0].(int)}
	})

	// Create a file
	file := lfs.New(LeagueFileType, 2023)

	// Write data
	fmt.Println("Writing file...")
	ts1, err := lfs.Write(file, []byte(`{"name": "Premier League", "teams": 20}`))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created version: %s\n", ts1)

	// Write another version
	fmt.Println("\nWriting another version...")
	ts2, err := lfs.Write(file, []byte(`{"name": "Premier League", "teams": 20, "year": 2023}`))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created version: %s\n", ts2)

	// Read a specific version
	fmt.Println("\nReading first version...")
	data, err := lfs.Read(file, ts1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Data: %s\n", string(data))

	// List all versions
	fmt.Println("\nListing all versions...")
	versions, err := lfs.Versions(file)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d versions:\n", len(versions))
	for i, v := range versions {
		fmt.Printf("  %d. %s (timestamp: %s)\n", i+1, v.LongString(), v)
	}

	// Get the latest version
	fmt.Println("\nGetting latest version...")
	latest, err := lfs.LastVersion(file)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Latest version: %s\n", latest)

	// Check if file has versions
	fmt.Println("\nChecking if file has versions...")
	hasVersions, err := lfs.HasSome(file)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Has versions: %v\n", hasVersions)

	// Remove a version
	fmt.Println("\nRemoving first version...")
	err = lfs.Remove(file, ts1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Version removed")

	// List versions again
	fmt.Println("\nListing versions after removal...")
	versions, err = lfs.Versions(file)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d versions:\n", len(versions))
	for i, v := range versions {
		fmt.Printf("  %d. %s\n", i+1, v.LongString())
	}

	fmt.Println("\n✓ Example completed successfully!")
}
