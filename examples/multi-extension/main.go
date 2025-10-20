// Example demonstrating multi-part file extensions (e.g., csv.gz)
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ericsperano/localfs"
)

// Define file types
const (
	ThemesFileType localfs.FileType = iota
	PlayersFileType
)

// ThemesFile has a multi-part extension (csv.gz)
type ThemesFile struct{}

func (f ThemesFile) Dir() string {
	return "catalog"
}

func (f ThemesFile) Name() string {
	return "themes"
}

func (f ThemesFile) Ext() string {
	return "csv.gz" // Multi-part extension
}

// PlayersFile also has a multi-part extension
type PlayersFile struct {
	season int
}

func (f PlayersFile) Dir() string {
	return fmt.Sprintf("%d/players", f.season)
}

func (f PlayersFile) Name() string {
	return "players"
}

func (f PlayersFile) Ext() string {
	return "json.gz" // Another multi-part extension
}

func main() {
	// Create a temporary directory for this example
	tmpDir, err := os.MkdirTemp("", "localfs-multiext-*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	fmt.Printf("Using directory: %s\n\n", tmpDir)

	// Create LocalFS instance
	lfs := localfs.New(tmpDir)

	// Register file types
	lfs.RegisterFileType(ThemesFileType, func(args ...any) localfs.File {
		return ThemesFile{}
	})
	lfs.RegisterFileType(PlayersFileType, func(args ...any) localfs.File {
		return PlayersFile{season: args[0].(int)}
	})

	// Create themes file (csv.gz extension)
	fmt.Println("Creating themes file with .csv.gz extension...")
	themesFile := lfs.New(ThemesFileType)
	ts1, err := lfs.Write(themesFile, []byte("id,name,year\n1,Castle,1978\n2,Space,1979"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created: catalog/themes.csv.gz.%s\n", ts1)

	// Create another version
	ts2, err := lfs.Write(themesFile, []byte("id,name,year\n1,Castle,1978\n2,Space,1979\n3,Pirates,1989"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created: catalog/themes.csv.gz.%s\n", ts2)

	// Create players file (json.gz extension)
	fmt.Println("\nCreating players file with .json.gz extension...")
	playersFile := lfs.New(PlayersFileType, 2023)
	ts3, err := lfs.Write(playersFile, []byte(`{"players":[{"id":1,"name":"Player A"}]}`))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created: 2023/players/players.json.gz.%s\n", ts3)

	// List versions for themes
	fmt.Println("\nListing all versions of themes file...")
	versions, err := lfs.Versions(themesFile)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d version(s):\n", len(versions))
	for i, v := range versions {
		fmt.Printf("  %d. themes.csv.gz.%s\n", i+1, v)
	}

	// Demonstrate Detect with multi-part extensions
	fmt.Println("\n=== DETECT WITH MULTI-PART EXTENSIONS ===")

	// Correct filename
	filename1 := fmt.Sprintf("themes.csv.gz.%s", ts1)
	fmt.Printf("\nDetecting: %s\n", filename1)
	detectedTS, err := lfs.Detect(filename1, themesFile)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✓ Detected! Timestamp: %s\n", detectedTS)
	}

	// Wrong extension (only .csv instead of .csv.gz)
	filename2 := fmt.Sprintf("themes.csv.%s", ts1)
	fmt.Printf("\nDetecting: %s\n", filename2)
	_, err = lfs.Detect(filename2, themesFile)
	if err != nil {
		fmt.Printf("❌ Not detected (expected): %v\n", err)
	}

	// Wrong extension (.gz.csv instead of .csv.gz)
	filename3 := fmt.Sprintf("themes.gz.csv.%s", ts1)
	fmt.Printf("\nDetecting: %s\n", filename3)
	_, err = lfs.Detect(filename3, themesFile)
	if err != nil {
		fmt.Printf("❌ Not detected (expected): %v\n", err)
	}

	// Demonstrate Find with multi-part extensions
	fmt.Println("\n=== FIND WITH MULTI-PART EXTENSIONS ===")

	fmt.Println("\nFinding all themes files...")
	timestamps, err := lfs.Find("catalog", themesFile)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d file(s):\n", len(timestamps))
	for i, ts := range timestamps {
		fmt.Printf("  %d. themes.csv.gz.%s\n", i+1, ts)
		data, _ := lfs.Read(themesFile, ts)
		fmt.Printf("     Preview: %s...\n", string(data[:min(50, len(data))]))
	}

	fmt.Println("\nFinding all players files...")
	timestamps, err = lfs.Find("2023/players", playersFile)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d file(s):\n", len(timestamps))
	for i, ts := range timestamps {
		fmt.Printf("  %d. players.json.gz.%s\n", i+1, ts)
	}

	// Read specific version
	fmt.Println("\nReading specific version...")
	data, err := lfs.Read(themesFile, ts2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Content of themes.csv.gz.%s:\n%s\n", ts2, string(data))

	fmt.Println("\n✓ Example completed successfully!")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
