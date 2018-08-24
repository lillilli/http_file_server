package types

import "time"

const (
	// TypeFile - тип объекта фс
	TypeFile = "file"
)

// FileInfo - структура информации о файле
type FileInfo struct {
	Name        string     `json:"name,omitempty"`
	Size        int64      `json:"size,omitempty"`
	URL         string     `json:"url_name,omitempty"`
	Type        string     `json:"type,omitempty"`
	Hash        string     `json:"hash,omitempty"`
	CreatedAt   *time.Time `json:"mtime,omitempty"`
	RequestedAt *time.Time `json:"atime,omitempty"`
	ModifiedAt  *time.Time `json:"ctime,omitempty"`
}
