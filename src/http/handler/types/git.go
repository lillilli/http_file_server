package types

import (
	"io"

	"github.com/lillilli/dr_web_exercise/src/types"
)

// Git интерфейс работы с репозиторием Git.
type Git interface {
	CloneOrFetch(repoPath, repoURL, tag string) error

	ComputeHash([]byte) string
	ComputeReaderHash(reader io.Reader) (string, error)

	CreateManifestForFiles(
		repoPath string,
		revision string,
		fileFilter func(string) bool,
	) (types.Manifest, error)

	LoadCommitFiles(
		repoPath, repoURL, revision string,
		fileNames []string,
		reader func(string, io.Reader) error,
	) error

	ResolveRevision(repoPath, ref string) (string, error)
}
