package versioned

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/lillilli/dr_web_exercise/src/types"
)

const (
	extensionSeparator = "."
)

func (h Handler) getSavedManifestHashs(project string, hash string) map[string]bool {
	savedManifests := make(map[string]bool)
	manifestDir := filepath.Dir(h.path.GetManifestCachePath(project, hash))

	if err := h.fs.MkdirAll(manifestDir, 0744); err != nil {
		h.log.Warnf("can`t create dir for savig manifect: %v", err)
		return savedManifests
	}

	fileNames, err := h.fs.Ls(manifestDir)
	if err != nil {
		h.log.Warnf("can`t get saved manifests: %v", err)
		return savedManifests
	}

	for _, fn := range fileNames {
		savedManifests[fn[:strings.LastIndex(fn, extensionSeparator)]] = true
	}

	return savedManifests
}

func (h Handler) assignManifestWithSavedManifest(manifest types.Manifest, savedManifestPath string) (types.Manifest, error) {
	savedManifest := make(types.Manifest)

	b, err := h.fs.ReadFile(savedManifestPath)
	if err != nil {
		h.log.Warnf("can`t read saved manifest: %s", err.Error())
		return savedManifest, err
	}

	if err := json.Unmarshal(b, &savedManifest); err != nil {
		h.log.Warnf("can`t unmarshal manifest, wrong manifest data: %s", err.Error())
		_ = h.fs.Remove(savedManifestPath)
		return savedManifest, err
	}

	return manifest.Asssign(savedManifest), nil
}

func (h Handler) saveManifestToFileSystem(project string, hash string, manifest types.Manifest) {
	b, err := json.Marshal(manifest)
	if err != nil {
		h.log.Warnf("can`t marshalize manifect for saving: %s", err.Error())
		return
	}

	err = h.fs.WriteFile(h.path.GetManifestCachePath(project, hash), b, 0644)
	if err != nil {
		h.log.Warnf("can`t write file: %v", err)
	}
}
