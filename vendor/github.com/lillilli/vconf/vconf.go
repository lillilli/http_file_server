package vconf

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Init - инициализация конфига, проставляются значения по-умолчанию для типов,
// или значения из тэга "default", или значения из env
func Init(cfg interface{}) error {
	return load("", cfg)
}

// InitFromFile - инициализация конфига, проставляются значения по-умолчанию для типов,
// или значения из тэга "default", или значения из env, или значение из конфига
func InitFromFile(path string, cfg interface{}) error {
	return load(path, cfg)
}

// load - получение конфига
func load(path string, config interface{}) error {
	setDefaultValues("", reflect.ValueOf(config))
	viper.AutomaticEnv()

	if path != "" {
		viper.SetConfigFile(path)

		if err := viper.ReadInConfig(); err != nil {
			return errors.Wrap(err, "unable to read config")
		}
	}

	if err := viper.Unmarshal(config); err != nil {
		return errors.Wrap(err, "unable to parse config")
	}

	return nil
}

// setDefaultValues - проставление значений по умолчанию в теге "default" для сруктуры
func setDefaultValues(prefix string, val reflect.Value) {
	val = valueOfInterfaceOrPtr(val)

	for i := 0; i < val.NumField(); i++ {
		valueField := valueOfInterfaceOrPtr(val.Field(i))
		typeField := val.Type().Field(i)
		nameField := structFieldName(prefix, typeField.Name)

		if valueField.Kind() == reflect.Struct {
			setDefaultValues(nameField, valueField)
			continue
		}

		if typeField.Tag.Get("default") != "" {
			viper.SetDefault(nameField, typeField.Tag.Get("default"))
		}
	}
}

func valueOfInterfaceOrPtr(val reflect.Value) reflect.Value {
	if val.Kind() == reflect.Interface && !val.IsNil() || val.Kind() == reflect.Ptr {
		elm := val.Elem()

		if elm.Kind() == reflect.Ptr && !elm.IsNil() && elm.Elem().Kind() == reflect.Ptr || elm.Kind() == reflect.Struct {
			val = elm
		}
	}

	return val
}

func structFieldName(structName string, fieldName string) (name string) {
	if structName != "" {
		name = fmt.Sprintf("%s.%s", structName, fieldName)
	} else {
		name = fieldName
	}

	return name
}
