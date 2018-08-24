package files

import (
	"bytes"
)

// PreFileUploadModifier - тип модификатора-функции, изменяющего файл перед записью на диск
type PreFileUploadModifier func(*bytes.Buffer) error

// PostFileUploadModifier - тип модификатора-функции, изменяющего файл после записи на диск
type PostFileUploadModifier func(string) error

// UploadModifiers - структура модификаторов загруженных файлов
type UploadModifiers struct {
	PostCallbackEnabled bool
	PreCallbackEnabled  bool

	PostCallback PostFileUploadModifier
	PreCallback  PreFileUploadModifier
}

// UploadModifier - модификатор загруженного файла
type UploadModifier func(*UploadModifiers)

func newModifiers(opt []UploadModifier) *UploadModifiers {
	opts := &UploadModifiers{}

	for _, o := range opt {
		o(opts)
	}

	return opts
}

// PreUploadModifier - модификатор, изменяющий файл перед его записью на диск
func PreUploadModifier(f PreFileUploadModifier) UploadModifier {
	return func(o *UploadModifiers) {
		o.PreCallbackEnabled = true
		o.PreCallback = f
	}
}

// PostUploadModifier - модификатор, изменяющий файл после его записи на диск
func PostUploadModifier(f PostFileUploadModifier) UploadModifier {
	return func(o *UploadModifiers) {
		o.PostCallbackEnabled = true
		o.PostCallback = f
	}
}
