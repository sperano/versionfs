// Basic example demonstrating core VersionFS functionality
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sperano/versionfs"
)

// Define file types
const (
	LeagueFileType versionfs.FileType = iota
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
	tmpDir, err := os.MkdirTemp("", "versionfs-example-*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	fmt.Printf("Using directory: %s\n\n", tmpDir)

	// Create VersionFS instance
	vfs := versionfs.New(tmpDir)

	// Register the file type
	vfs.RegisterFileType(LeagueFileType, func(args ...any) versionfs.File {
		return LeagueFile{season: args[0].(int)}
	})

	// Create a file
	file := vfs.New(LeagueFileType, 2023)

	// Write data
	fmt.Println("Writing file...")
	ts1, err := vfs.Write(file, []byte(`{"name": "Premier League", "teams": 20}`))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created version: %s\n", ts1)

	// Write another version
	fmt.Println("\nWriting another version...")
	ts2, err := vfs.Write(file, []byte(`{"name": "Premier League", "teams": 20, "year": 2023}`))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created version: %s\n", ts2)

	// Read a specific version
	fmt.Println("\nReading first version...")
	data, err := vfs.Read(file, ts1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Data: %s\n", string(data))

	// List all versions
	fmt.Println("\nListing all versions...")
	versions, err := vfs.Versions(file)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d versions:\n", len(versions))
	for i, v := range versions {
		fmt.Printf("  %d. %s (timestamp: %s)\n", i+1, v.LongString(), v)
	}

	// Get the latest version
	fmt.Println("\nGetting latest version...")
	latest, err := vfs.LastVersion(file)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Latest version: %s\n", latest)

	// Check if file has versions
	fmt.Println("\nChecking if file has versions...")
	hasVersions, err := vfs.HasSome(file)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Has versions: %v\n", hasVersions)

	// Remove a version
	fmt.Println("\nRemoving first version...")
	err = vfs.Remove(file, ts1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Version removed")

	// List versions again
	fmt.Println("\nListing versions after removal...")
	versions, err = vfs.Versions(file)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d versions:\n", len(versions))
	for i, v := range versions {
		fmt.Printf("  %d. %s\n", i+1, v.LongString())
	}

	fmt.Println("\n✓ Example completed successfully!")
}
