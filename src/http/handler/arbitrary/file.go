package arbitrary

import "github.com/lillilli/dr_web_exercise/src/http/handler/types"

// GenerateUploadedFileInfo - генерация структуры загруженного файла
func GenerateUploadedFileInfo(filename, hash string) map[string]types.FileInfo {
	info := make(map[string]types.FileInfo)
	info[filename] = types.FileInfo{Hash: hash}

	return info
}
