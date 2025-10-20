# localfs

A Go library for managing versioned files in a local filesystem with automatic timestamping.

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org/doc/install)
[![Coverage](https://img.shields.io/badge/coverage-93.3%25-brightgreen.svg)](https://github.com/ericsperano/localfs)

## Features

- **Automatic Versioning**: Every file write creates a new timestamped version
- **Type-Safe File Operations**: Register custom file types with constructors
- **Multi-Part Extensions**: Support for extensions like `csv.gz`
- **File Detection**: Check if filenames match specific file type patterns
- **Directory Scanning**: Find all files of a specific type in a directory
- **Version Management**: List, read, and remove specific versions
- **Clean API**: Returns empty results instead of errors for missing directories

## Installation

```bash
go get github.com/ericsperano/localfs
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/ericsperano/localfs"
)

// Define your file types
const (
    LeagueFileType localfs.FileType = iota
    RosterFileType
)

// Implement the File interface
type LeagueFile struct {
    season int
}

func (f LeagueFile) Dir() string  { return fmt.Sprintf("%d/league", f.season) }
func (f LeagueFile) Name() string { return "league" }
func (f LeagueFile) Ext() string  { return "json" }

func main() {
    // Create a new LocalFS instance
    lfs := localfs.New("./data")

    // Register file types
    lfs.RegisterFileType(LeagueFileType, func(args ...any) localfs.File {
        return LeagueFile{season: args[0].(int)}
    })

    // Create a file
    file := lfs.New(LeagueFileType, 2023)

    // Write data (creates: 2023/league/league.json.20231019140523)
    ts, err := lfs.Write(file, []byte(`{"name": "Premier League"}`))
    if err != nil {
        panic(err)
    }
    fmt.Printf("Created version: %s\n", ts)

    // Read the file
    data, err := lfs.Read(file, ts)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Data: %s\n", data)

    // List all versions
    versions, err := lfs.Versions(file)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Found %d versions\n", len(versions))

    // Get the latest version
    latest, err := lfs.LastVersion(file)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Latest: %s\n", latest)
}
```

## File Naming Convention

Files follow the pattern: `name.extension.timestamp`

Examples:
- `league.json.20231019140523`
- `themes.csv.gz.20231019140523`
- `roster-12-2023-10-19.json.20231019140523`

The timestamp format is: `YYYYMMDDHHmmss` (e.g., `20231019140523` = October 19, 2023, 14:05:23)

## API Reference

### Core Operations

#### Write
```go
func (l *LocalFS) Write(file File, data []byte) (Timestamp, error)
```
Writes data to a file and returns the generated timestamp. Creates the directory if it doesn't exist.

#### Read
```go
func (l *LocalFS) Read(file File, ts Timestamp) ([]byte, error)
```
Reads a specific version of a file.

#### Remove
```go
func (l *LocalFS) Remove(file File, ts Timestamp) error
```
Removes a specific version of a file.

### Version Management

#### Versions
```go
func (l *LocalFS) Versions(file File) ([]Timestamp, error)
```
Lists all versions of a file, sorted newest first. Returns empty slice if directory doesn't exist.

#### LastVersion
```go
func (l *LocalFS) LastVersion(file File) (Timestamp, error)
```
Returns the most recent version of a file. Returns `ErrNoVersions` if no versions exist.

#### HasSome
```go
func (l *LocalFS) HasSome(file File) (bool, error)
```
Checks if any versions of a file exist.

### File Type Operations

#### Detect (Detector)
```go
func (l *LocalFS) Detect(filename string, file File) (Timestamp, error)
```
Checks if a filename matches a file type pattern and extracts the timestamp.

**Example:**
```go
file := lfs.New(LeagueFileType, 2023)
ts, err := lfs.Detect("league.json.20231019140523", file)
if err != nil {
    // Filename doesn't match the pattern
    fmt.Println("Not a league file")
} else {
    fmt.Printf("Found league file with timestamp: %s\n", ts)
}
```

#### Find (Finder)
```go
func (l *LocalFS) Find(dir string, file File) ([]Timestamp, error)
```
Searches a directory for all files matching a file type, returning their timestamps.

**Example:**
```go
file := lfs.New(LeagueFileType, 2023)
timestamps, err := lfs.Find("2023/league", file)
if err != nil {
    panic(err)
}
for _, ts := range timestamps {
    fmt.Printf("Found version: %s\n", ts)
}
```

### Utility Functions

#### PathExists
```go
func (l *LocalFS) PathExists(path string) (bool, error)
```
Checks if a path exists in the filesystem.

#### MkdirAll
```go
func (l *LocalFS) MkdirAll(path string, perm os.FileMode) error
```
Creates a directory and all parent directories.

## File Interface

Implement the `File` interface for your custom file types:

```go
type File interface {
    Dir() string   // Directory path (relative to RootPath)
    Name() string  // Base filename without extension
    Ext() string   // File extension (can be multi-part like "csv.gz")
}
```

## Timestamp Helpers

```go
// Create from time.Time
ts := localfs.NewFromTime(time.Now())

// Parse from string
ts, err := localfs.NewTimestamp("20231019140523")

// Parse simple date format
ts, err := localfs.NewTimestampSimple("2023-10-19")

// Format timestamps
fmt.Println(ts.String())            // "20231019140523"
fmt.Println(ts.LongString())        // "2023-10-19 14:05:23"
fmt.Println(ts.SimpleDateString())  // "2023-10-19"
fmt.Println(ts.Time())              // time.Time object
```

## Examples

### Multi-Part Extensions

```go
type ThemesFile struct{}

func (f ThemesFile) Dir() string  { return "catalog" }
func (f ThemesFile) Name() string { return "themes" }
func (f ThemesFile) Ext() string  { return "csv.gz" }  // Multi-part extension

// Creates: catalog/themes.csv.gz.20231019140523
```

### Parameterized File Types

```go
type RosterFile struct {
    season int
    teamID int
    date   string
}

func (f RosterFile) Dir() string {
    return fmt.Sprintf("%d/roster/team-%d", f.season, f.teamID)
}

func (f RosterFile) Name() string {
    return fmt.Sprintf("roster-%d-%s", f.teamID, f.date)
}

func (f RosterFile) Ext() string {
    return "json"
}

// Creates: 2023/roster/team-12/roster-12-2023-10-19.json.20231019140523
```

## Design Goals

**LocalFS** is designed for managing different types of files that:
- Have no parameters (e.g., Lego Themes catalog)
- Have parameters (e.g., Team Roster with team ID and date)
- Need multiple generations/versions
- Require automatic versioning without database overhead

### Planned Features
- ❌ **Differs**: Compare two files of the same type (not yet implemented)

## Testing

```bash
# Run tests
go test

# Run tests with coverage
go test -cover

# Run tests verbosely
go test -v
```

Current test coverage: **93.3%**

## Requirements

- Go 1.21 or higher
- Dependencies:
  - `github.com/rs/zerolog` - Logging
  - `github.com/golang-module/carbon/v2` - Time utilities (indirect)

## License

See LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
