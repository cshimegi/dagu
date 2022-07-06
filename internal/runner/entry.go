package runner

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/yohamta/dagu/internal/admin"
	"github.com/yohamta/dagu/internal/config"
	"github.com/yohamta/dagu/internal/utils"
)

type Entry struct {
	Next time.Time
	Job  Job
}

type EntryReader interface {
	Read(now time.Time) ([]*Entry, error)
}

type entryReader struct {
	Admin *admin.Config
}

var _ EntryReader = (*entryReader)(nil)

func (er *entryReader) Read(now time.Time) ([]*Entry, error) {
	cl := config.Loader{}
	entries := []*Entry{}
	for {
		fis, err := os.ReadDir(er.Admin.DAGs)
		if err != nil {
			return nil, fmt.Errorf("failed to read entries directory: %w", err)
		}
		for _, fi := range fis {
			if utils.MatchExtension(fi.Name(), config.EXTENSIONS) {
				dag, err := cl.LoadHeadOnly(
					filepath.Join(er.Admin.DAGs, fi.Name()),
				)
				if err != nil {
					log.Printf("failed to read dag config: %s", err)
					continue
				}
				for _, sc := range dag.Schedule {
					next := sc.Next(now)
					entries = append(entries, &Entry{
						Next: sc.Next(now),
						Job: &job{
							DAG:       dag,
							Config:    er.Admin,
							StartTime: next,
						},
					})
				}
			}
		}
		return entries, nil
	}
}
