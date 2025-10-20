// Package versionfs provides a versioned file system for managing files with automatic timestamping.
//
// VersionFS allows you to store and manage different types of files with automatic version control.
// Each file write creates a new timestamped version, and you can list, read, and manage versions.
//
// Example usage:
//
//	vfs := versionfs.New("./data")
//	vfs.RegisterFileType(LeagueFileType, func(args ...any) versionfs.File {
//	    return LeagueFile{season: args[0].(int)}
//	})
//	file := vfs.New(LeagueFileType, 2023)
//	ts, err := vfs.Write(file, []byte("data"))
//
// Files are stored with the pattern: dir/name.ext.timestamp
// For example: 2023/league/league.json.20231019140523
package versionfs

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	path_ "path"
	"sort"
	"strings"
	"time"
)

// FileType represents a type of file in the system.
// Users should define their own FileType constants using iota.
type FileType int

// File defines the interface that all file types must implement.
// It specifies the directory, base name, and extension for a file.
type File interface {
	// Dir returns the directory path relative to the VersionFS root.
	// Example: "2023/league" or "catalog"
	Dir() string

	// Name returns the base filename without extension or timestamp.
	// Example: "league" or "roster-12-2023-10-19"
	Name() string

	// Ext returns the file extension without the dot.
	// Can be multi-part for compressed files.
	// Examples: "json", "txt", "csv.gz"
	Ext() string
}

// Path constructs the full file path for a given file and timestamp.
// Returns a path in the format: dir/name.ext.timestamp
//
// Example: "2023/league/league.json.20231019140523"
func Path(file File, version Timestamp) string {
	return fmt.Sprintf("%s/%s.%s.%s", file.Dir(), file.Name(), file.Ext(), version)
}

// Constructor is a function type for creating File instances.
// It accepts variadic arguments to support parameterized file types.
type Constructor func(args ...any) File

// VersionFS manages versioned files in a local filesystem.
// It maintains a root path and a registry of file type constructors.
type VersionFS struct {
	// RootPath is the base directory for all file operations.
	RootPath string
	// constructors maps FileType to their constructor functions.
	constructors map[FileType]Constructor
}

// New creates a new VersionFS instance with the specified root path.
// The root path is where all files will be stored.
//
// Example:
//
//	vfs := versionfs.New("./data")
func New(rootPath string) *VersionFS {
	return &VersionFS{
		RootPath:     rootPath,
		constructors: make(map[FileType]Constructor),
	}
}

// RegisterFileType registers a constructor function for a file type.
// The constructor will be called when creating new instances of this file type.
//
// Example:
//
//	lfs.RegisterFileType(LeagueFileType, func(args ...any) localfs.File {
//	    return LeagueFile{season: args[0].(int)}
//	})
func (v *VersionFS) RegisterFileType(ftype FileType, constructor Constructor) {
	v.constructors[ftype] = constructor
}

// Write writes data to a file and returns the generated timestamp.
// The file is created with the pattern: dir/name.ext.timestamp
// The directory is created automatically if it doesn't exist.
//
// Example:
//
//	ts, err := lfs.Write(file, []byte("data"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Created version: %s\n", ts)
func (v *VersionFS) Write(file File, data []byte) (Timestamp, error) {
	log.Debug().Msgf("Writing file %s/%s.%s.?", file.Dir(), file.Name(), file.Ext())
	if err := v.MkdirAll(file.Dir(), 0755); err != nil {
		return Timestamp{}, err
	}
	ts := NewFromTime(time.Now())
	filepath := Path(file, ts)
	return ts, os.WriteFile(path_.Join(v.RootPath, filepath), data, 0644)
}

// Read reads a specific version of a file identified by its timestamp.
// Returns an error if the file doesn't exist.
//
// Example:
//
//	data, err := lfs.Read(file, timestamp)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (v *VersionFS) Read(file File, ts Timestamp) ([]byte, error) {
	log.Debug().Msgf("Reading file %s/%s.%s.%s", file.Dir(), file.Name(), file.Ext(), ts)
	return os.ReadFile(path_.Join(v.RootPath, Path(file, ts)))
}

// Remove deletes a specific version of a file identified by its timestamp.
// Returns an error if the file doesn't exist or cannot be deleted.
//
// Example:
//
//	err := lfs.Remove(file, timestamp)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (v *VersionFS) Remove(file File, ts Timestamp) error {
	log.Debug().Msgf("remove file %s/%s.%s.%s", file.Dir(), file.Name(), file.Ext(), ts)
	return os.Remove(path_.Join(v.RootPath, Path(file, ts)))
}

// New creates a new File instance using a registered constructor.
// Panics if the file type has not been registered.
//
// Example:
//
//	file := lfs.New(LeagueFileType, 2023)
func (v *VersionFS) New(ftype FileType, args ...any) File {
	c, ok := v.constructors[ftype]
	if !ok {
		panic(fmt.Errorf("file type %d not registered", ftype))
	}
	return c(args...)
}

// ErrNoVersions is returned when no versions of a file exist.
var ErrNoVersions = errors.New("no version found")

// HasSome checks if any versions of a file exist.
// Returns true if at least one version exists, false otherwise.
//
// Example:
//
//	exists, err := lfs.HasSome(file)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if exists {
//	    fmt.Println("File has versions")
//	}
func (v *VersionFS) HasSome(file File) (bool, error) {
	versions, err := v.Versions(file)
	if err != nil {
		return false, err
	}
	return len(versions) > 0, nil
}

// LastVersion returns the most recent version (timestamp) of a file.
// Returns ErrNoVersions if no versions exist.
// Returns an empty slice if the directory doesn't exist.
//
// Example:
//
//	latest, err := lfs.LastVersion(file)
//	if err == localfs.ErrNoVersions {
//	    fmt.Println("No versions found")
//	} else if err != nil {
//	    log.Fatal(err)
//	}
func (v *VersionFS) LastVersion(file File) (Timestamp, error) {
	versions, err := v.Versions(file)
	if err != nil {
		return Timestamp{}, err
	}
	if len(versions) == 0 {
		return Timestamp{}, ErrNoVersions
	}
	return versions[0], nil
}

// Versions returns all versions (timestamps) of a file, sorted newest first.
// Returns an empty slice if the directory doesn't exist or contains no matching files.
// Only returns versions for files that match the exact name and extension.
//
// Example:
//
//	versions, err := lfs.Versions(file)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, ts := range versions {
//	    fmt.Printf("Version: %s\n", ts)
//	}
func (v *VersionFS) Versions(file File) ([]Timestamp, error) {
	entries, err := os.ReadDir(path_.Join(v.RootPath, file.Dir()))
	if err != nil {
		if os.IsNotExist(err) {
			return []Timestamp{}, nil
		}
		return nil, err
	}
	var versions []Timestamp
	fname := file.Name()
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].Name() > entries[j].Name()
	})
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), fname) { // AND extension
			rest := entry.Name()[len(fname):]
			// next char has to be a dot
			if len(rest) == 0 || !strings.HasPrefix(rest, ".") {
				log.Warn().Msgf("unexpected file: %s/%s", file.Dir(), entry.Name())
				continue
			}
			rest = rest[1:]
			tokens := strings.Split(rest, ".")
			ts, err := NewTimestamp(tokens[len(tokens)-1])
			if err != nil {
				log.Warn().Msgf("unexpected timestamp for file: %s/%s", file.Dir(), entry.Name())
				continue
			}
			versions = append(versions, ts)
		}
	}
	return versions, nil
}

// Detect checks if a filename matches the given file type pattern and extracts the timestamp.
// Returns the timestamp if the filename matches, or an error describing why it doesn't match.
// Validates that the filename has the correct name, extension, and timestamp format.
//
// Expected filename format: name.ext.timestamp or name.ext1.ext2.timestamp
//
// Example:
//
//	ts, err := lfs.Detect("league.json.20231019140523", file)
//	if err != nil {
//	    fmt.Println("Not a league file")
//	} else {
//	    fmt.Printf("Found version: %s\n", ts)
//	}
func (v *VersionFS) Detect(filename string, file File) (Timestamp, error) {
	fname := file.Name()
	fext := file.Ext()

	// Check if filename starts with the file name
	if !strings.HasPrefix(filename, fname) {
		return Timestamp{}, fmt.Errorf("filename %q does not match file name %q", filename, fname)
	}

	rest := filename[len(fname):]

	// Next char must be a dot
	if len(rest) == 0 || !strings.HasPrefix(rest, ".") {
		return Timestamp{}, fmt.Errorf("filename %q has invalid format, expected dot after name", filename)
	}

	rest = rest[1:] // Remove the dot
	tokens := strings.Split(rest, ".")

	// Expected format: name.ext.timestamp or name.ext1.ext2.timestamp
	// We need at least extension.timestamp
	if len(tokens) < 2 {
		return Timestamp{}, fmt.Errorf("filename %q has invalid format, expected ext.timestamp", filename)
	}

	// Check if extension matches (handle multi-part extensions like csv.gz)
	// Join all tokens except the last one (which should be timestamp)
	actualExt := strings.Join(tokens[:len(tokens)-1], ".")
	if actualExt != fext {
		return Timestamp{}, fmt.Errorf("filename %q has extension %q but expected %q", filename, actualExt, fext)
	}

	// Last token should be the timestamp
	ts, err := NewTimestamp(tokens[len(tokens)-1])
	if err != nil {
		return Timestamp{}, fmt.Errorf("filename %q has invalid timestamp: %w", filename, err)
	}

	return ts, nil
}

// Find searches a directory for all files matching the given file type.
// Returns a list of timestamps for files that match the file's name and extension, sorted newest first.
// Returns an empty slice if the directory doesn't exist or contains no matching files.
// Skips files with invalid timestamps or incorrect extensions.
//
// Example:
//
//	timestamps, err := lfs.Find("2023/league", file)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, ts := range timestamps {
//	    fmt.Printf("Found version: %s\n", ts)
//	    data, _ := lfs.Read(file, ts)
//	    // process data...
//	}
func (v *VersionFS) Find(dir string, file File) ([]Timestamp, error) {
	entries, err := os.ReadDir(path_.Join(v.RootPath, dir))
	if err != nil {
		if os.IsNotExist(err) {
			return []Timestamp{}, nil
		}
		return nil, err
	}

	var results []Timestamp
	fname := file.Name()
	fext := file.Ext()

	// Sort by name descending (newest first)
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].Name() > entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Check if filename starts with the file name
		if !strings.HasPrefix(entry.Name(), fname) {
			continue
		}

		rest := entry.Name()[len(fname):]

		// Next char must be a dot
		if len(rest) == 0 || !strings.HasPrefix(rest, ".") {
			continue
		}

		rest = rest[1:] // Remove the dot
		tokens := strings.Split(rest, ".")

		// Expected format: name.ext.timestamp or name.ext1.ext2.timestamp
		// We need at least extension.timestamp
		if len(tokens) < 2 {
			continue
		}

		// Check if extension matches (handle multi-part extensions like csv.gz)
		// Join all tokens except the last one (which should be timestamp)
		actualExt := strings.Join(tokens[:len(tokens)-1], ".")
		if actualExt != fext {
			continue
		}

		// Last token should be the timestamp
		ts, err := NewTimestamp(tokens[len(tokens)-1])
		if err != nil {
			log.Warn().Msgf("unexpected timestamp for file: %s/%s", dir, entry.Name())
			continue
		}

		results = append(results, ts)
	}

	return results, nil
}

// PathExists checks if a path exists in the filesystem.
// Returns true if the path exists, false if it doesn't exist.
// Returns an error for other filesystem errors (e.g., permission denied).
//
// Example:
//
//	exists, err := lfs.PathExists("2023/league")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if exists {
//	    fmt.Println("Directory exists")
//	}
func (v *VersionFS) PathExists(path string) (bool, error) {
	_, err := os.Stat(path_.Join(v.RootPath, path))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// MkdirAll creates a directory and all necessary parent directories.
// The path is relative to the VersionFS root path.
// Does nothing if the directory already exists.
//
// Example:
//
//	err := lfs.MkdirAll("2023/league", 0755)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (v *VersionFS) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path_.Join(v.RootPath, path), perm)
}
