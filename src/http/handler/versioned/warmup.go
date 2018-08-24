package versioned

import (
	"io"
	"sync"

	"github.com/pkg/errors"
	"github.com/lillilli/dr_web_exercise/src/types"
)

func (h Handler) warmUpManifest(project string, manifest types.Manifest) {
	preparedManifest := manifest.GroupByCommits()

	wg := &sync.WaitGroup{}
	wg.Add(len(preparedManifest))

	for rev, files := range preparedManifest {
		go h.warmUpCommitFiles(wg, project, rev, files)
	}

	wg.Wait()
}

func (h Handler) warmUpCommitFiles(wg *sync.WaitGroup, project, rev string, files []string) {
	repoPath := h.path.GetProjectRepoStorageDir(project)
	repoUrl := h.cfg.Projects[project].Repository
	_ = h.git.LoadCommitFiles(repoPath, repoUrl, rev, files, func(filename string, reader io.Reader) error {
		cacheFilename := h.path.GetFileCachePath(project, filename)
		if err := h.fs.WriteReader(cacheFilename, reader, 0644); err != nil {
			return errors.Wrap(err, "copy file to cache path failed")
		}

		return nil
	})

	wg.Done()
}
