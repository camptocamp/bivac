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
	Engine    string `label:"engine" ini:"engine" config:"Engine"`
	NoVerify  bool   `label:"no_verify" ini:"no_verify" config:"NoVerify"`
	Ignore    bool   `label:"ignore" ini:"ignore" default:"false"`
	TargetURL string `label:"target_url" ini:"target_url" config:"TargetURL"`

	Duplicity struct {
		FullIfOlderThan string `label:"full_if_older_than" ini:"full_if_older_than" config:"FullIfOlderThan"`
		RemoveOlderThan string `label:"remove_older_than" ini:"remove_older_than" config:"RemoveOlderThan"`
	} `label:"duplicity" ini:"duplicity" config:"Duplicity"`

	RClone struct {
	} `label:"rclone" ini:"rclone" config:"RClone"`
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
	ref := ptrRef.Elem()

	return v.getStructConfig(ref, c, iniOverrides, "")
}

func (v *Volume) getStructConfig(ref reflect.Value, c *config.Config, iniOverrides *ini.File, structTag reflect.StructTag) error {
	if ref.Kind() != reflect.Struct {
		return fmt.Errorf("not a struct ptr")
	}

	refType := ref.Type()
	for i := 0; i < refType.NumField(); i++ {
		if ref.Field(i).Kind() == reflect.Struct {
			log.Debugf("Found a struct %s", refType.Field(i).Name)
			err := v.getStructConfig(ref.Field(i), c, iniOverrides, refType.Field(i).Tag)
			if err != nil {
				return err
			}
			continue
		}

		log.Debugf("Getting value for field %s", refType.Field(i).Name)
		value := v.getField(refType.Field(i), c, iniOverrides, structTag)
		log.Debugf("Got volume config: %s=%v", refType.Field(i).Name, value)
		if err := setField(ref.Field(i), refType.Field(i), value); err != nil {
			return err
		}
	}
	return nil
}

func (v *Volume) getField(field reflect.StructField, c *config.Config, iniOverrides *ini.File, structTag reflect.StructTag) (value interface{}) {
	log.Debugf("Getting field %s from docker label", field.Name)
	var label string
	if prefix := structTag.Get("label"); prefix != "" {
		label = fmt.Sprintf("%s.", prefix)
	}
	label += field.Tag.Get("label")
	value, _ = util.GetVolumeLabel(v.Volume, label)
	if value != "" {
		return
	}

	log.Debugf("Getting field %s from ini overrides", field.Name)
	value = getIniValue(iniOverrides, structTag.Get("ini"), field.Tag.Get("ini"))
	if value != "" {
		return
	}

	log.Debugf("Getting field %s from general config", field.Name)
	value = getConfigKey(c, structTag.Get("config"), field.Tag.Get("config"))
	if value != nil {
		return
	}

	value = field.Tag.Get("default")
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
	if key == "" {
		return nil
	}
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
