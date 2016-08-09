package volume

import (
	"fmt"
	"os"
	"reflect"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/conplicity/config"
	"github.com/camptocamp/conplicity/util"
	"github.com/docker/engine-api/types"
	"github.com/go-ini/ini"
)

// Volume provides backup methods for a single Docker volume
type Volume struct {
	*types.Volume
	Target    string
	BackupDir string
	Mount     string
	Config    *Config
}

// Config is the volume's configuration parameters
type Config struct {
	Engine   string `label:"engine" ini:"engine" config:"Engine"`
	NoVerify bool   `label:"no_verify" ini:"no_verify" config:"NoVerify"`

	// Duplicity config
	FullIfOlderThan string `label:"full_if_older_than" ini-section:"duplicity" ini:"duplicity.full_if_older_than" config-section:"Duplicity" config:"FullIfOlderThan"`
	RemoveOlderThan string `label:"remove_older_than" ini-section:"duplicity" ini:"duplicity.remove_older_than" config-section:"Duplicity" config:"RemoveOlderThan"`
}

// NewVolume returns a new Volume for a given types.Volume struct
func NewVolume(v *types.Volume, c *config.Config) *Volume {
	vol := &Volume{
		Volume: v,
		Config: &Config{},
	}

	err := vol.getConfig(c)
	if err != nil {
		log.Fatal(err)
	}

	return vol
}

func (v *Volume) getConfig(c *config.Config) error {
	var iniOverrides *ini.File
	overridesFile := v.Mountpoint + "/.conplicity.overrides"
	if f, err := os.Stat(overridesFile); err == nil && f.Mode().IsRegular() {
		iniOverrides, err = ini.Load(overridesFile)
		if err != nil {
			return fmt.Errorf("could not read overrides file %s: %v", overridesFile, err)
		}
	}

	ptrRef := reflect.ValueOf(v.Config)
	if ptrRef.Kind() != reflect.Ptr {
		return fmt.Errorf("not a ptr")
	}
	ref := ptrRef.Elem()
	if ref.Kind() != reflect.Struct {
		return fmt.Errorf("not a struct ptr")
	}

	refType := ref.Type()
	for i := 0; i < refType.NumField(); i++ {
		value := v.getField(refType.Field(i), c, iniOverrides)
		log.Debugf("Got volume config: %s=%v", refType.Field(i).Name, value)
		if err := setField(ref.Field(i), refType.Field(i), value); err != nil {
			return err
		}
	}
	return nil
}

func (v *Volume) getField(field reflect.StructField, c *config.Config, iniOverrides *ini.File) (value interface{}) {
	log.Debugf("Getting field %s from docker label", field.Name)
	value, _ = util.GetVolumeLabel(v.Volume, field.Tag.Get("label"))
	if value != "" {
		return
	}

	log.Debugf("Getting field %s from ini overrides", field.Name)
	value = getIniValue(iniOverrides, field.Tag.Get("ini-section"), field.Tag.Get("ini"))
	if value != "" {
		return
	}

	log.Debugf("Getting field %s from general config", field.Name)
	value = getConfigKey(c, field.Tag.Get("config-section"), field.Tag.Get("config"))

	return
}

func getIniValue(file *ini.File, section, key string) (value string) {
	if file == nil {
		return ""
	}
	val, err := file.Section(section).GetKey(key)
	if err == nil {
		value = val.String()
	}
	return
}

func getConfigKey(s interface{}, section, key string) interface{} {
	r := reflect.ValueOf(s)
	var f reflect.Value
	if section == "" {
		f = reflect.Indirect(r).FieldByName(key)
	} else {
		s := reflect.Indirect(r).FieldByName(section)
		f = s.FieldByName(key)
	}
	return f.Interface()
}

func setField(field reflect.Value, fieldType reflect.StructField, value interface{}) error {
	v := reflect.ValueOf(value)

	log.Debugf("[%s] field: %s, value:%s", fieldType.Name, field.Kind(), v.Kind())

	if field.Kind() == v.Kind() {
		log.Debugf("field and value have the same kind: %v", field.Kind())
		field.Set(v)
		return nil
	}

	if v.Kind() != reflect.String {
		return fmt.Errorf("mismatched value type and do not know how to convert from type %s", v.Kind())
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value.(string))
	case reflect.Bool:
		bvalue, err := strconv.ParseBool(value.(string))
		if err != nil {
			return err
		}
		field.SetBool(bvalue)
	default:
		return fmt.Errorf("unsupported type %s", field.Kind())
	}
	return nil
}
