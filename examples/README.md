# LocalFS Examples

This directory contains runnable examples demonstrating various features of the LocalFS library.

## Running the Examples

Each example is a standalone Go program. To run an example:

```bash
cd examples/basic
go run main.go
```

Or from the project root:

```bash
go run ./examples/basic
```

## Available Examples

### 1. Basic Usage (`basic/`)

Demonstrates core LocalFS functionality:
- Creating a LocalFS instance
- Registering file types
- Writing files (creates new versions)
- Reading specific versions
- Listing all versions
- Getting the latest version
- Checking if versions exist
- Removing versions

**Run:**
```bash
go run ./examples/basic
```

### 2. Detect and Find (`detect-find/`)

Demonstrates file detection and searching:
- Using `Detect()` to check if filenames match file types
- Using `Find()` to search directories for matching files
- Handling multiple file types in the same directory
- Working with parameterized file types (e.g., team rosters)
- Error handling for non-matching filenames

**Run:**
```bash
go run ./examples/detect-find
```

### 3. Multi-Part Extensions (`multi-extension/`)

Demonstrates handling of multi-part file extensions:
- Creating files with extensions like `.csv.gz` and `.json.gz`
- Verifying that extension matching is exact
- Using `Detect()` with multi-part extensions
- Using `Find()` with multi-part extensions
- Reading and managing compressed file versions

**Run:**
```bash
go run ./examples/multi-extension
```

## Example Output

All examples create temporary directories for demonstration purposes and clean up after themselves. The output shows:
- File paths being created
- Timestamps generated
- Version listings
- Detection results
- File content previews

## Creating Your Own Examples

To create a custom example:

1. Implement the `File` interface for your file type:
```go
type MyFile struct {
    // your fields
}

func (f MyFile) Dir() string  { return "path/to/dir" }
func (f MyFile) Name() string { return "filename" }
func (f MyFile) Ext() string  { return "ext" }
```

2. Register your file type:
```go
lfs.RegisterFileType(MyFileType, func(args ...any) localfs.File {
    return MyFile{ /* initialize from args */ }
})
```

3. Use the LocalFS methods to manage your files!

## Tips

- Use temporary directories (`os.MkdirTemp`) for testing
- Always clean up with `defer os.RemoveAll(tmpDir)`
- Check for errors - the examples show proper error handling patterns
- Use descriptive file type constants with `iota`
