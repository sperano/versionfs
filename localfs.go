package localfs

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	path_ "path"
	"sort"
	"strings"
	"time"
)

type FileType int

type File interface {
	Dir() string
	Name() string
	Ext() string
}

func Path(file File, version Timestamp) string {
	return fmt.Sprintf("%s/%s.%s.%s", file.Dir(), file.Name(), file.Ext(), version)
}

type Constructor func(args ...any) File

type LocalFS struct {
	RootPath     string
	constructors map[FileType]Constructor
}

func New(rootPath string) *LocalFS {
	return &LocalFS{
		RootPath:     rootPath,
		constructors: make(map[FileType]Constructor),
	}
}

func (l *LocalFS) RegisterFileType(ftype FileType, constructor Constructor) {
	l.constructors[ftype] = constructor
}

func (l *LocalFS) Write(file File, data []byte) (Timestamp, error) {
	log.Info().Msgf("write file %s/%s.?.%s", file.Dir(), file.Name(), file.Ext())
	if err := os.MkdirAll(path_.Join(l.RootPath, file.Dir()), 0755); err != nil {
		return Timestamp{}, err
	}
	ts := NewFromTime(time.Now())
	filepath := Path(file, ts)
	return ts, os.WriteFile(path_.Join(l.RootPath, filepath), data, 0644)
}

func (l *LocalFS) Read(file File, ts Timestamp) ([]byte, error) {
	log.Info().Msgf("read file %s/%s.%s.%s", file.Dir(), file.Name(), ts, file.Ext())
	return os.ReadFile(path_.Join(l.RootPath, Path(file, ts)))
}

func (l *LocalFS) Remove(file File, ts Timestamp) error {
	log.Info().Msgf("remove file %s/%s.%s.%s", file.Dir(), file.Name(), ts, file.Ext())
	return os.Remove(path_.Join(l.RootPath, Path(file, ts)))
}

func (l *LocalFS) New(ftype FileType, args ...any) File {
	c, ok := l.constructors[ftype]
	if !ok {
		panic(fmt.Errorf("file type %d not registered", ftype))
	}
	return c(args...)
}

func (l *LocalFS) Versions(file File) ([]Timestamp, error) {
	entries, err := os.ReadDir(path_.Join(l.RootPath, file.Dir()))
	if err != nil {
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

func (l *LocalFS) PathExists(path string) (bool, error) {
	_, err := os.Stat(path_.Join(l.RootPath, path))
	if err != nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (l *LocalFS) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path_.Join(l.RootPath, path), perm)
}
