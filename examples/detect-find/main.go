// Example demonstrating Detect and Find functionality
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
	RosterFileType
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

// RosterFile implements the File interface with parameters
type RosterFile struct {
	season int
	teamID int
}

func (f RosterFile) Dir() string {
	return fmt.Sprintf("%d/rosters", f.season)
}

func (f RosterFile) Name() string {
	return fmt.Sprintf("roster-%d", f.teamID)
}

func (f RosterFile) Ext() string {
	return "json"
}

func main() {
	// Create a temporary directory for this example
	tmpDir, err := os.MkdirTemp("", "versionfs-detect-find-*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	fmt.Printf("Using directory: %s\n\n", tmpDir)

	// Create VersionFS instance
	vfs := versionfs.New(tmpDir)

	// Register file types
	vfs.RegisterFileType(LeagueFileType, func(args ...any) versionfs.File {
		return LeagueFile{season: args[0].(int)}
	})
	vfs.RegisterFileType(RosterFileType, func(args ...any) versionfs.File {
		return RosterFile{season: args[0].(int), teamID: args[1].(int)}
	})

	// Create some league files
	fmt.Println("Creating league files...")
	leagueFile := vfs.New(LeagueFileType, 2023)
	ts1, _ := vfs.Write(leagueFile, []byte(`{"name": "Premier League"}`))
	ts2, _ := vfs.Write(leagueFile, []byte(`{"name": "Premier League", "updated": true}`))
	fmt.Printf("Created 2 league file versions: %s, %s\n", ts1, ts2)

	// Create some roster files in the same season
	fmt.Println("\nCreating roster files...")
	roster1 := vfs.New(RosterFileType, 2023, 1)
	roster2 := vfs.New(RosterFileType, 2023, 2)
	vfs.Write(roster1, []byte(`{"team": 1, "players": ["Player A", "Player B"]}`))
	vfs.Write(roster2, []byte(`{"team": 2, "players": ["Player C", "Player D"]}`))
	fmt.Println("Created 2 roster files")

	// Demonstrate Detect functionality
	fmt.Println("\n=== DETECT FUNCTIONALITY ===")

	// Test valid filename
	filename := fmt.Sprintf("league.json.%s", ts1)
	fmt.Printf("\nDetecting filename: %s\n", filename)
	detectedTS, err := vfs.Detect(filename, leagueFile)
	if err != nil {
		fmt.Printf("❌ Not detected: %v\n", err)
	} else {
		fmt.Printf("✓ Detected! Timestamp: %s\n", detectedTS)
	}

	// Test invalid filename (wrong name)
	invalidFilename := fmt.Sprintf("roster-1.json.%s", ts1)
	fmt.Printf("\nDetecting filename: %s\n", invalidFilename)
	_, err = vfs.Detect(invalidFilename, leagueFile)
	if err != nil {
		fmt.Printf("❌ Not detected (expected): %v\n", err)
	}

	// Test invalid filename (wrong extension)
	invalidExt := fmt.Sprintf("league.txt.%s", ts1)
	fmt.Printf("\nDetecting filename: %s\n", invalidExt)
	_, err = vfs.Detect(invalidExt, leagueFile)
	if err != nil {
		fmt.Printf("❌ Not detected (expected): %v\n", err)
	}

	// Demonstrate Find functionality
	fmt.Println("\n=== FIND FUNCTIONALITY ===")

	// Find all league files
	fmt.Println("\nFinding all league files in '2023/league'...")
	timestamps, err := vfs.Find("2023/league", leagueFile)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d league file(s):\n", len(timestamps))
	for i, ts := range timestamps {
		fmt.Printf("  %d. %s (%s)\n", i+1, ts, ts.LongString())
		// Read and display the content
		data, _ := vfs.Read(leagueFile, ts)
		fmt.Printf("     Content: %s\n", string(data))
	}

	// Find all roster files
	fmt.Println("\nFinding all roster files in '2023/rosters'...")
	roster1Timestamps, err := vfs.Find("2023/rosters", roster1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d roster file(s) for team 1:\n", len(roster1Timestamps))
	for i, ts := range roster1Timestamps {
		fmt.Printf("  %d. %s\n", i+1, ts.LongString())
	}

	roster2Timestamps, err := vfs.Find("2023/rosters", roster2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d roster file(s) for team 2:\n", len(roster2Timestamps))
	for i, ts := range roster2Timestamps {
		fmt.Printf("  %d. %s\n", i+1, ts.LongString())
	}

	// Find in non-existent directory
	fmt.Println("\nFinding in non-existent directory '2024/league'...")
	emptyResults, err := vfs.Find("2024/league", leagueFile)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d file(s) (expected 0)\n", len(emptyResults))

	fmt.Println("\n✓ Example completed successfully!")
}
