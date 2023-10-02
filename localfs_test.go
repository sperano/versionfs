package localfs

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

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

func newTestLocalFS() *LocalFS {
	lfs := New("./test-data/")
	lfs.RegisterFileType(LeagueFileType, func(args ...any) File {
		return fileLeague{season: args[0].(int)}
	})
	lfs.RegisterFileType(RosterFileType, func(args ...any) File {
		return fileRoster{season: args[0].(int), teamID: args[1].(int), date: args[2].(string)}
	})
	return lfs
}

func newTmpLocalFS(t *testing.T) (string, *LocalFS) {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	lfs := New(dir)
	lfs.RegisterFileType(LeagueFileType, func(args ...any) File {
		return fileLeague{season: args[0].(int)}
	})
	return dir, lfs
}

func TestLocalFS_New(t *testing.T) {
	t.Parallel()
	lfs := newTestLocalFS()
	file := lfs.New(LeagueFileType, 2023)
	const TS = "20211125011947"
	ts, _ := NewTimestamp(TS)
	assert.Equal(t, fmt.Sprintf("2023/league/league.%s.txt", TS), Path(file, ts))
	file = lfs.New(RosterFileType, 2023, 3, "2023-10-19")
	assert.Equal(t, fmt.Sprintf("2023/roster/team-3/roster-3-2023-10-19.%s.json", TS), Path(file, ts))
	// should panic if not registered
	assert.Panics(t, func() { lfs.New(99) }, "The code did not panic")
}

func TestLocalFS_Read(t *testing.T) {
	t.Parallel()
	lfs := newTestLocalFS()
	file := lfs.New(LeagueFileType, 2023)
	//file := newLeague(2023)
	ts, _ := NewTimestamp("20211125011947")
	data, err := lfs.Read(file, ts)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "hello world 2\n", string(data))
}

func TestLocalFS_Versions(t *testing.T) {
	t.Parallel()
	lfs := newTestLocalFS()
	file := lfs.New(LeagueFileType, 2023)
	versions, err := lfs.Versions(file)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 3, len(versions))
	assert.Equal(t, "20211218030527", versions[0].String())
	assert.Equal(t, "20211125011947", versions[1].String())
	assert.Equal(t, "20211125011946", versions[2].String())
}

func TestLocalFS_Versions_Error(t *testing.T) {
	t.Parallel()
	lfs := newTestLocalFS()
	lfs.rootPath = "./test-data/missing"
	file := lfs.New(LeagueFileType, 2023)
	versions, err := lfs.Versions(file)
	assert.Nil(t, versions)
	assert.Equal(t, "open test-data/missing/2023/league: no such file or directory", err.Error())
}

func TestLocalFS_Write(t *testing.T) {
	t.Parallel()
	dir, lfs := newTmpLocalFS(t)
	defer func() { _ = os.RemoveAll(dir) }()
	file := lfs.New(LeagueFileType, 2023)
	ts, err := lfs.Write(file, []byte("new hello world"))
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path.Join(lfs.rootPath, Path(file, ts)))
	assert.Equal(t, "new hello world", string(data))
}

//func TestLocalFS_Write_Error(t *testing.T) {
//	t.Parallel()
//}

func TestLocalFS_Remove(t *testing.T) {
	t.Parallel()
	dir, lfs := newTmpLocalFS(t)
	defer func() { _ = os.RemoveAll(dir) }()
	file := lfs.New(LeagueFileType, 2023)
	ts, err := lfs.Write(file, []byte("new hello world"))
	if err != nil {
		t.Fatal(err)
	}
	finfo, err := os.Stat(path.Join(lfs.rootPath, Path(file, ts)))
	assert.NotNil(t, finfo)
	assert.Nil(t, err)
	if err := lfs.Remove(file, ts); err != nil {
		t.Fatal(err)
	}
	finfo, err = os.Stat(path.Join(lfs.rootPath, Path(file, ts)))
	assert.Nil(t, finfo)
	assert.True(t, errors.Is(err, os.ErrNotExist))
}

func TestLocalFS_Remove_Err(t *testing.T) {
	t.Parallel()
	lfs := newTestLocalFS()
	lfs.rootPath = "./test-data/missing"
	file := lfs.New(LeagueFileType, 2023)
	ts, _ := NewTimestamp("20211125011947")
	if err := lfs.Remove(file, ts); err != nil {
		assert.True(t, errors.Is(err, os.ErrNotExist))
	}
}
