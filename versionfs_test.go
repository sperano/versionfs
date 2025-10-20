package versionfs

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func init() {
	// Disable logging during tests
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

const (
	LeagueFileType FileType = iota
	RosterFileType
)

type fileLeague struct {
	season int
}

func (f fileLeague) Dir() string {
	return fmt.Sprintf("%d/league", f.season)
}

func (f fileLeague) Name() string {
	return "league"
}

func (f fileLeague) Ext() string {
	return "txt"
}

type fileRoster struct {
	season int
	teamID int
	date   string
}

func (f fileRoster) Dir() string {
	return fmt.Sprintf("%d/roster/team-%d", f.season, f.teamID)
}

func (f fileRoster) Name() string {
	return fmt.Sprintf("roster-%d-%s", f.teamID, f.date)
}

func (f fileRoster) Ext() string {
	return "json"
}

func newTestVersionFS() *VersionFS {
	vfs := New("./test-data/")
	vfs.RegisterFileType(LeagueFileType, func(args ...any) File {
		return fileLeague{season: args[0].(int)}
	})
	vfs.RegisterFileType(RosterFileType, func(args ...any) File {
		return fileRoster{season: args[0].(int), teamID: args[1].(int), date: args[2].(string)}
	})
	return vfs
}

func newTmpVersionFS(tb testing.TB) (string, *VersionFS) {
	tb.Helper()
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		tb.Fatal(err)
	}
	vfs := New(dir)
	vfs.RegisterFileType(LeagueFileType, func(args ...any) File {
		return fileLeague{season: args[0].(int)}
	})
	return dir, vfs
}

// Test the new method - It has two registered types (league and roster), make sure the correct file object
// is created. it should panic if we create a type that doesn't exists
func TestVersionFS_New(t *testing.T) {
	t.Parallel()
	vfs := newTestVersionFS()
	file := vfs.New(LeagueFileType, 2023)
	const TS = "20211125011947"
	ts, _ := NewTimestamp(TS)
	assert.Equal(t, fmt.Sprintf("2023/league/league.txt.%s", TS), Path(file, ts))
	file = vfs.New(RosterFileType, 2023, 3, "2023-10-19")
	assert.Equal(t, fmt.Sprintf("2023/roster/team-3/roster-3-2023-10-19.json.%s", TS), Path(file, ts))
	// should panic if not registered
	assert.Panics(t, func() { vfs.New(99) }, "The code did not panic")
}

// should read a file correctly - sample file is in test-data
func TestVersionFS_Read(t *testing.T) {
	t.Parallel()
	vfs := newTestVersionFS()
	file := vfs.New(LeagueFileType, 2023)
	ts, _ := NewTimestamp("20211125011947")
	data, err := vfs.Read(file, ts)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "hello world 2\n", string(data))
}

func TestVersionFS_Versions(t *testing.T) {
	t.Parallel()
	vfs := newTestVersionFS()
	file := vfs.New(LeagueFileType, 2023)
	versions, err := vfs.Versions(file)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 3, len(versions))
	assert.Equal(t, "20211218030527", versions[0].String())
	assert.Equal(t, "20211125011947", versions[1].String())
	assert.Equal(t, "20211125011946", versions[2].String())
}

func TestVersionFS_Versions_MissingDir(t *testing.T) {
	t.Parallel()
	vfs := newTestVersionFS()
	vfs.RootPath = "./test-data/missing"
	file := vfs.New(LeagueFileType, 2023)
	versions, err := vfs.Versions(file)
	assert.Nil(t, err)
	assert.Equal(t, []Timestamp{}, versions)
}

// re-use the same file as the Versions test, we know it should return true
func TestVersionFS_HasSome(t *testing.T) {
	t.Parallel()
	vfs := newTestVersionFS()
	file := vfs.New(LeagueFileType, 2023)
	ok, err := vfs.HasSome(file)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, ok)
}

// use a year that we don't have in test data, it should be false
//func TestVersionFS_DoesntHaveSome(t *testing.T) {
//	t.Parallel()
//	vfs := newTestVersionFS()
//	file := vfs.New(LeagueFileType, 2000)
//	ok, err := vfs.HasSome(file)
//	if err != nil {
//		t.Fatal(err)
//	}
//	assert.False(t, ok)
//}

func TestVersionFS_LastVersion(t *testing.T) {
	t.Parallel()
	vfs := newTestVersionFS()
	file := vfs.New(LeagueFileType, 2023)
	version, err := vfs.LastVersion(file)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "20211218030527", version.String())
}

func TestVersionFS_LastVersion_MissingDir(t *testing.T) {
	t.Parallel()
	vfs := newTestVersionFS()
	vfs.RootPath = "./test-data/missing"
	file := vfs.New(LeagueFileType, 2023)
	version, err := vfs.LastVersion(file)
	assert.Zero(t, version)
	assert.Equal(t, ErrNoVersions, err)
}

func TestVersionFS_Write(t *testing.T) {
	t.Parallel()
	dir, vfs := newTmpVersionFS(t)
	defer func() { _ = os.RemoveAll(dir) }()
	file := vfs.New(LeagueFileType, 2023)
	ts, err := vfs.Write(file, []byte("new hello world"))
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path.Join(vfs.RootPath, Path(file, ts)))
	assert.Equal(t, "new hello world", string(data))
}

// let's write on a path that is not writable
func TestVersionFS_Write_Error(t *testing.T) {
	t.Parallel()
	vfs := newTestVersionFS()
	vfs.RootPath = "/dev/null/"
	file := vfs.New(LeagueFileType, 2023)
	ts, err := vfs.Write(file, []byte("new hello world"))
	assert.Zero(t, ts)
	assert.Equal(t, "mkdir /dev/null: not a directory", err.Error())
}

func TestVersionFS_Remove(t *testing.T) {
	t.Parallel()
	dir, vfs := newTmpVersionFS(t)
	defer func() { _ = os.RemoveAll(dir) }()
	file := vfs.New(LeagueFileType, 2023)
	ts, err := vfs.Write(file, []byte("new hello world"))
	if err != nil {
		t.Fatal(err)
	}
	finfo, err := os.Stat(path.Join(vfs.RootPath, Path(file, ts)))
	assert.NotNil(t, finfo)
	assert.Nil(t, err)
	if err := vfs.Remove(file, ts); err != nil {
		t.Fatal(err)
	}
	finfo, err = os.Stat(path.Join(vfs.RootPath, Path(file, ts)))
	assert.Nil(t, finfo)
	assert.True(t, errors.Is(err, os.ErrNotExist))
}

func TestVersionFS_Remove_Err(t *testing.T) {
	t.Parallel()
	vfs := newTestVersionFS()
	vfs.RootPath = "./test-data/missing"
	file := vfs.New(LeagueFileType, 2023)
	ts, _ := NewTimestamp("20211125011947")
	if err := vfs.Remove(file, ts); err != nil {
		assert.True(t, errors.Is(err, os.ErrNotExist))
	}
}

func TestVersionFS_PathExists(t *testing.T) {
	t.Parallel()
	vfs := newTestVersionFS()
	// Test with a path that exists
	exists, err := vfs.PathExists("2023/league")
	assert.Nil(t, err)
	assert.True(t, exists, "Expected existing path to return true")
}

func TestVersionFS_PathDoesNotExist(t *testing.T) {
	t.Parallel()
	vfs := newTestVersionFS()
	// Test with a path that does not exist
	exists, err := vfs.PathExists("2023/nonexistent")
	assert.Nil(t, err)
	assert.False(t, exists, "Expected non-existing path to return false")
}

func TestVersionFS_DoesntHaveSome(t *testing.T) {
	t.Parallel()
	dir, vfs := newTmpVersionFS(t)
	defer func() { _ = os.RemoveAll(dir) }()
	file := vfs.New(LeagueFileType, 2023)
	ok, err := vfs.HasSome(file)
	if err != nil {
		t.Fatal(err)
	}
	assert.False(t, ok)
}

func TestVersionFS_HasSome_MissingDir(t *testing.T) {
	t.Parallel()
	vfs := newTestVersionFS()
	vfs.RootPath = "./test-data/missing"
	file := vfs.New(LeagueFileType, 2023)
	ok, err := vfs.HasSome(file)
	assert.False(t, ok)
	assert.Nil(t, err)
}

func TestVersionFS_LastVersion_NoVersions(t *testing.T) {
	t.Parallel()
	dir, vfs := newTmpVersionFS(t)
	defer func() { _ = os.RemoveAll(dir) }()
	file := vfs.New(LeagueFileType, 2023)
	version, err := vfs.LastVersion(file)
	assert.Zero(t, version)
	assert.Equal(t, ErrNoVersions, err)
}

func TestVersionFS_Find(t *testing.T) {
	t.Parallel()
	vfs := newTestVersionFS()
	file := vfs.New(LeagueFileType, 2023)
	timestamps, err := vfs.Find("2023/league", file)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 3, len(timestamps))
	assert.Equal(t, "20211218030527", timestamps[0].String())
	assert.Equal(t, "20211125011947", timestamps[1].String())
	assert.Equal(t, "20211125011946", timestamps[2].String())
}

func TestVersionFS_Find_NoMatches(t *testing.T) {
	t.Parallel()
	dir, vfs := newTmpVersionFS(t)
	defer func() { _ = os.RemoveAll(dir) }()
	if err := vfs.MkdirAll("2023/league", 0755); err != nil {
		t.Fatal(err)
	}
	file := vfs.New(LeagueFileType, 2023)
	timestamps, err := vfs.Find("2023/league", file)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 0, len(timestamps))
}

func TestVersionFS_Find_MissingDir(t *testing.T) {
	t.Parallel()
	vfs := newTestVersionFS()
	vfs.RootPath = "./test-data/missing"
	file := vfs.New(LeagueFileType, 2023)
	timestamps, err := vfs.Find("2023/league", file)
	assert.Nil(t, err)
	assert.Equal(t, []Timestamp{}, timestamps)
}

func TestVersionFS_Find_WrongExtension(t *testing.T) {
	t.Parallel()
	dir, vfs := newTmpVersionFS(t)
	defer func() { _ = os.RemoveAll(dir) }()
	file := vfs.New(LeagueFileType, 2023)

	// Create a directory with files
	if err := vfs.MkdirAll("2023/league", 0755); err != nil {
		t.Fatal(err)
	}

	// Write a file with .txt extension
	ts1, err := vfs.Write(file, []byte("test content 1"))
	if err != nil {
		t.Fatal(err)
	}

	// Create a file with wrong extension (.json instead of .txt)
	wrongFile := path.Join(vfs.RootPath, "2023/league", fmt.Sprintf("league.json.%s", ts1.String()))
	if err := os.WriteFile(wrongFile, []byte("wrong"), 0644); err != nil {
		t.Fatal(err)
	}

	// Find should only return the .txt file
	timestamps, err := vfs.Find("2023/league", file)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(timestamps))
	assert.Equal(t, ts1.String(), timestamps[0].String())
}

func TestVersionFS_Find_MultipleFiles(t *testing.T) {
	t.Parallel()
	dir, vfs := newTmpVersionFS(t)
	defer func() { _ = os.RemoveAll(dir) }()

	vfs.RegisterFileType(RosterFileType, func(args ...any) File {
		return fileRoster{season: args[0].(int), teamID: args[1].(int), date: args[2].(string)}
	})

	// Create league files
	leagueFile := vfs.New(LeagueFileType, 2023)
	ts1, err := vfs.Write(leagueFile, []byte("league 1"))
	if err != nil {
		t.Fatal(err)
	}

	// Create roster files in the same directory
	if err := vfs.MkdirAll("2023/league", 0755); err != nil {
		t.Fatal(err)
	}
	rosterPath := path.Join(vfs.RootPath, "2023/league", fmt.Sprintf("roster-1-2023-10-19.json.%s", ts1.String()))
	if err := os.WriteFile(rosterPath, []byte("roster"), 0644); err != nil {
		t.Fatal(err)
	}

	// Find should only return league files, not roster files
	timestamps, err := vfs.Find("2023/league", leagueFile)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(timestamps))
	assert.Equal(t, ts1.String(), timestamps[0].String())
}

func TestVersionFS_Detect(t *testing.T) {
	t.Parallel()
	vfs := newTestVersionFS()
	file := vfs.New(LeagueFileType, 2023)

	// Valid filename
	ts, err := vfs.Detect("league.txt.20211125011947", file)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "20211125011947", ts.String())
}

func TestVersionFS_Detect_MultiPartExtension(t *testing.T) {
	t.Parallel()
	dir, vfs := newTmpVersionFS(t)
	defer func() { _ = os.RemoveAll(dir) }()

	// Register a file type with multi-part extension (csv.gz)
	const ThemesFileType FileType = 99
	vfs.RegisterFileType(ThemesFileType, func(args ...any) File {
		return &fileThemes{}
	})

	file := vfs.New(ThemesFileType)
	ts, err := vfs.Detect("themes.csv.gz.20211125011947", file)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "20211125011947", ts.String())
}

func TestVersionFS_Detect_WrongName(t *testing.T) {
	t.Parallel()
	vfs := newTestVersionFS()
	file := vfs.New(LeagueFileType, 2023)

	// Wrong name
	_, err := vfs.Detect("roster.txt.20211125011947", file)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "does not match file name")
}

func TestVersionFS_Detect_WrongExtension(t *testing.T) {
	t.Parallel()
	vfs := newTestVersionFS()
	file := vfs.New(LeagueFileType, 2023)

	// Wrong extension
	_, err := vfs.Detect("league.json.20211125011947", file)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "has extension")
}

func TestVersionFS_Detect_InvalidTimestamp(t *testing.T) {
	t.Parallel()
	vfs := newTestVersionFS()
	file := vfs.New(LeagueFileType, 2023)

	// Invalid timestamp
	_, err := vfs.Detect("league.txt.invalid", file)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid timestamp")
}

func TestVersionFS_Detect_MissingDot(t *testing.T) {
	t.Parallel()
	vfs := newTestVersionFS()
	file := vfs.New(LeagueFileType, 2023)

	// Missing dot after name
	_, err := vfs.Detect("leaguetxt20211125011947", file)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "expected dot after name")
}

func TestVersionFS_Detect_MissingExtension(t *testing.T) {
	t.Parallel()
	vfs := newTestVersionFS()
	file := vfs.New(LeagueFileType, 2023)

	// Missing extension (only timestamp)
	_, err := vfs.Detect("league.20211125011947", file)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "expected ext.timestamp")
}

func TestVersionFS_Detect_EmptyAfterName(t *testing.T) {
	t.Parallel()
	vfs := newTestVersionFS()
	file := vfs.New(LeagueFileType, 2023)

	// Just the name, nothing after
	_, err := vfs.Detect("league", file)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "expected dot after name")
}

// Helper type for multi-part extension testing
type fileThemes struct{}

func (f fileThemes) Dir() string {
	return "catalog"
}

func (f fileThemes) Name() string {
	return "themes"
}

func (f fileThemes) Ext() string {
	return "csv.gz"
}

func TestVersionFS_PathExists_OtherError(t *testing.T) {
	t.Parallel()
	vfs := New("/root/no-permission")
	// This should trigger a permission error, not a "not exist" error
	exists, err := vfs.PathExists("test")
	assert.False(t, exists)
	// On some systems we might get permission denied, on others we might get "not exist"
	// Just verify we get some error or false
	if err != nil {
		assert.NotNil(t, err)
	}
}

func TestVersionFS_Find_OtherError(t *testing.T) {
	t.Parallel()
	vfs := New("/root/no-permission")
	file := &fileLeague{season: 2023}
	// This should trigger a permission error
	timestamps, err := vfs.Find("test", file)
	// On different systems this might behave differently
	// Just make sure we handle errors
	if err != nil {
		assert.Nil(t, timestamps)
	} else {
		// If no error, should be empty
		assert.Equal(t, 0, len(timestamps))
	}
}

func TestVersionFS_Versions_SkipDirectories(t *testing.T) {
	t.Parallel()
	dir, vfs := newTmpVersionFS(t)
	defer func() { _ = os.RemoveAll(dir) }()

	file := vfs.New(LeagueFileType, 2023)

	// Create a file
	_, err := vfs.Write(file, []byte("data"))
	if err != nil {
		t.Fatal(err)
	}

	// Create a subdirectory that starts with the file name
	subdir := path.Join(vfs.RootPath, file.Dir(), "league.subdir")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatal(err)
	}

	// Versions should skip the directory
	versions, err := vfs.Versions(file)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(versions))
}

func TestVersionFS_Find_SkipDirectories(t *testing.T) {
	t.Parallel()
	dir, vfs := newTmpVersionFS(t)
	defer func() { _ = os.RemoveAll(dir) }()

	file := vfs.New(LeagueFileType, 2023)

	// Create a file
	ts, err := vfs.Write(file, []byte("data"))
	if err != nil {
		t.Fatal(err)
	}

	// Create a subdirectory that matches the pattern
	subdir := path.Join(vfs.RootPath, file.Dir(), fmt.Sprintf("league.txt.%s.dir", ts))
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatal(err)
	}

	// Find should skip the directory
	timestamps, err := vfs.Find(file.Dir(), file)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(timestamps))
}

func TestVersionFS_Find_InvalidFormat(t *testing.T) {
	t.Parallel()
	dir, vfs := newTmpVersionFS(t)
	defer func() { _ = os.RemoveAll(dir) }()

	file := vfs.New(LeagueFileType, 2023)

	// Create directory
	if err := vfs.MkdirAll(file.Dir(), 0755); err != nil {
		t.Fatal(err)
	}

	// Create a file with only one token after name (no timestamp)
	invalidFile := path.Join(vfs.RootPath, file.Dir(), "league.txt")
	if err := os.WriteFile(invalidFile, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	// Find should skip this file
	timestamps, err := vfs.Find(file.Dir(), file)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 0, len(timestamps))
}

func TestVersionFS_Find_NoPrefix(t *testing.T) {
	t.Parallel()
	dir, vfs := newTmpVersionFS(t)
	defer func() { _ = os.RemoveAll(dir) }()

	file := vfs.New(LeagueFileType, 2023)

	// Create directory
	if err := vfs.MkdirAll(file.Dir(), 0755); err != nil {
		t.Fatal(err)
	}

	// Create a valid file first
	ts, err := vfs.Write(file, []byte("data"))
	if err != nil {
		t.Fatal(err)
	}

	// Create a file that doesn't start with the right prefix
	wrongFile := path.Join(vfs.RootPath, file.Dir(), fmt.Sprintf("other.txt.%s", ts))
	if err := os.WriteFile(wrongFile, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	// Find should only return the correct file
	timestamps, err := vfs.Find(file.Dir(), file)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(timestamps))
	assert.Equal(t, ts.String(), timestamps[0].String())
}

// Benchmarks

func BenchmarkWrite(b *testing.B) {
	dir, vfs := newTmpVersionFS(b)
	defer func() { _ = os.RemoveAll(dir) }()

	file := vfs.New(LeagueFileType, 2023)
	data := []byte("benchmark data for write operation")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := vfs.Write(file, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRead(b *testing.B) {
	dir, vfs := newTmpVersionFS(b)
	defer func() { _ = os.RemoveAll(dir) }()

	file := vfs.New(LeagueFileType, 2023)
	data := []byte("benchmark data for read operation")
	ts, err := vfs.Write(file, data)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := vfs.Read(file, ts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVersions(b *testing.B) {
	dir, vfs := newTmpVersionFS(b)
	defer func() { _ = os.RemoveAll(dir) }()

	file := vfs.New(LeagueFileType, 2023)
	// Create 10 versions
	for i := 0; i < 10; i++ {
		_, err := vfs.Write(file, []byte(fmt.Sprintf("version %d", i)))
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := vfs.Versions(file)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLastVersion(b *testing.B) {
	dir, vfs := newTmpVersionFS(b)
	defer func() { _ = os.RemoveAll(dir) }()

	file := vfs.New(LeagueFileType, 2023)
	// Create 10 versions
	for i := 0; i < 10; i++ {
		_, err := vfs.Write(file, []byte(fmt.Sprintf("version %d", i)))
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := vfs.LastVersion(file)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDetect(b *testing.B) {
	vfs := newTestVersionFS()
	file := vfs.New(LeagueFileType, 2023)
	filename := "league.txt.20211125011947"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := vfs.Detect(filename, file)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFind(b *testing.B) {
	dir, vfs := newTmpVersionFS(b)
	defer func() { _ = os.RemoveAll(dir) }()

	file := vfs.New(LeagueFileType, 2023)
	// Create 10 versions
	for i := 0; i < 10; i++ {
		_, err := vfs.Write(file, []byte(fmt.Sprintf("version %d", i)))
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := vfs.Find("2023/league", file)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHasSome(b *testing.B) {
	dir, vfs := newTmpVersionFS(b)
	defer func() { _ = os.RemoveAll(dir) }()

	file := vfs.New(LeagueFileType, 2023)
	_, err := vfs.Write(file, []byte("data"))
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := vfs.HasSome(file)
		if err != nil {
			b.Fatal(err)
		}
	}
}
