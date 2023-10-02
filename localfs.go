package localfs

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"path"
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
	return fmt.Sprintf("%s/%s.%s.%s", file.Dir(), file.Name(), version, file.Ext())
}

type Constructor func(args ...any) File

type LocalFS struct {
	rootPath     string
	constructors map[FileType]Constructor
}

func New(rootPath string) *LocalFS {
	return &LocalFS{
		rootPath:     rootPath,
		constructors: make(map[FileType]Constructor),
	}
}

func (l *LocalFS) RegisterFileType(ftype FileType, constructor Constructor) {
	l.constructors[ftype] = constructor
}

func (l *LocalFS) Write(file File, data []byte) (Timestamp, error) {
	log.Info().Msgf("write file %s/%s.?.%s", file.Dir(), file.Name(), file.Ext())
	if err := os.MkdirAll(path.Join(l.rootPath, file.Dir()), 0755); err != nil {
		return Timestamp{}, err
	}
	ts := NewFromTime(time.Now())
	filepath := Path(file, ts)
	return ts, os.WriteFile(path.Join(l.rootPath, filepath), data, 0644)
}

func (l *LocalFS) Read(file File, ts Timestamp) ([]byte, error) {
	log.Info().Msgf("read file %s/%s.%s.%s", file.Dir(), file.Name(), ts, file.Ext())
	return os.ReadFile(path.Join(l.rootPath, Path(file, ts)))
}

func (l *LocalFS) Remove(file File, ts Timestamp) error {
	log.Info().Msgf("remove file %s/%s.%s.%s", file.Dir(), file.Name(), ts, file.Ext())
	return os.Remove(path.Join(l.rootPath, Path(file, ts)))
}

func (l *LocalFS) New(ftype FileType, args ...any) File {
	c, ok := l.constructors[ftype]
	if !ok {
		panic(fmt.Errorf("file type %d not registered", ftype))
	}
	return c(args...)
}

func (l *LocalFS) Versions(file File) ([]Timestamp, error) {
	entries, err := os.ReadDir(path.Join(l.rootPath, file.Dir()))
	if err != nil {
		return nil, err
	}
	var versions []Timestamp
	fname := file.Name()
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].Name() > entries[j].Name()
	})
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), fname) {
			rest := entry.Name()[len(fname):]
			// next char has to be a dot
			if len(rest) == 0 || !strings.HasPrefix(rest, ".") {
				log.Warn().Msgf("unexpected file: %s/%s", file.Dir(), entry.Name())
				continue
			}
			rest = rest[1:]
			tokens := strings.Split(rest, ".")
			ts, err := NewTimestamp(tokens[0])
			if err != nil {
				log.Warn().Msgf("unexpected timestamp for file: %s/%s", file.Dir(), entry.Name())
				continue
			}
			versions = append(versions, ts)
		}
	}
	return versions, nil
}
