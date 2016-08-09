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
		if err := setField(ref.Field(i), refType.Field(i), value); err != nil {
			return err
		}
	}
	return nil
}

func (v *Volume) getField(field reflect.StructField, c *config.Config, iniOverrides *ini.File) string {
	log.Debugf("Attempting to get field from docker label")
	value, _ := util.GetVolumeLabel(v.Volume, field.Tag.Get("label"))
	if value == "" && iniOverrides != nil {
		log.Debugf("Attempting to get field from ini overrides")
		iniSection := field.Tag.Get("ini-section")
		iniKey := field.Tag.Get("ini")
		val, err := iniOverrides.Section(iniSection).GetKey(iniKey)
		if err == nil {
			value = val.String()
		}
	}
	if value == "" {
		log.Debugf("Attempting to get field from general config")
		confSection := field.Tag.Get("config-section")
		confKey := field.Tag.Get("config")
		value = getConfigKey(c, confSection, confKey)
	}
	log.Debugf("Volume config: %s=%s", field.Name, value)
	return value
}

func getConfigKey(s interface{}, section, key string) string {
	r := reflect.ValueOf(s)
	var f interface{}
	if section == "" {
		f = reflect.Indirect(r).FieldByName(key)
	} else {
		s := reflect.Indirect(r).FieldByName(section)
		f = s.FieldByName(key)
	}
	// FIXME: This is UGLY!
	return fmt.Sprintf("%v", f)
}

func setField(field reflect.Value, refType reflect.StructField, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Bool:
		bvalue, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(bvalue)
	default:
		return fmt.Errorf("unsupported type")
	}
	return nil
}
