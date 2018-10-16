package volume

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/bivac/config"
	"github.com/camptocamp/bivac/metrics"
	"github.com/camptocamp/bivac/util"
	"github.com/go-ini/ini"
)

// Volume provides backup methods for a single Docker volume
type Volume struct {
	ID             string
	Name           string
	BackupDir      string
	Mount          string
	Mountpoint     string
	Driver         string
	Labels         map[string]string
	LabelPrefix    string
	ReadOnly       bool
	Config         *Config
	MetricsHandler *metrics.PrometheusMetrics
	HostBind       string
	Hostname       string
	Namespace      string
}

// Config is the volume's configuration parameters
type Config struct {
	Engine          string `label:"engine" ini:"engine" config:"Engine"`
	NoVerify        bool   `label:"no_verify" ini:"no_verify" config:"NoVerify"`
	Ignore          bool   `label:"ignore" ini:"ignore" default:"false"`
	TargetURL       string `label:"target_url" ini:"target_url" config:"TargetURL"`
	RemoveOlderThan string `label:"remove_older_than" ini:"remove_older_than" config:"RemoveOlderThan"`
}

// MountedVolume stores mounted volumes inside a container
type MountedVolume struct {
	PodID       string
	ContainerID string
	HostID      string
	Volume      *Volume
	Path        string
}

// NewVolume returns a new Volume for a given types.Volume struct
func NewVolume(v *Volume, c *config.Config, h string) *Volume {
	err := v.getConfig(c)
	if err != nil {
		log.Fatal(err)
	}

	err = v.setupMetrics(c, h)
	if err != nil {
		log.Error(err)
	}

	return v
}

// LogTime adds a new metric even with the current time
func (v *Volume) LogTime(event string) (err error) {
	metricName := fmt.Sprintf("bivac_%s", event)
	startTimeMetric := v.MetricsHandler.NewMetric(metricName, "counter")
	err = startTimeMetric.UpdateEvent(
		&metrics.Event{
			Labels: map[string]string{
				"volume": v.Name,
			},
			Value: strconv.FormatInt(time.Now().Unix(), 10),
		},
	)
	if err != nil {
		return
	}
	err = v.MetricsHandler.Push()
	return
}

func (v *Volume) setupMetrics(c *config.Config, h string) (err error) {
	v.MetricsHandler = metrics.NewMetrics(h, v.Name, c.Metrics.PushgatewayURL)
	util.CheckErr(err, "Failed to set up metrics: %v", "fatal")
	return
}

func (v *Volume) getConfig(c *config.Config) error {
	var iniOverrides *ini.File
	overridesFile := v.Mountpoint + "/.bivac.overrides"
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
			//log.Debugf("Found a struct %s", refType.Field(i).Name)
			err := v.getStructConfig(ref.Field(i), c, iniOverrides, refType.Field(i).Tag)
			if err != nil {
				return err
			}
			continue
		}

		//log.Debugf("Getting value for field %s", refType.Field(i).Name)
		value := v.getField(refType.Field(i), c, iniOverrides, structTag)
		//log.Debugf("Got volume config: %s=%v", refType.Field(i).Name, value)
		if err := setField(ref.Field(i), refType.Field(i), value); err != nil {
			return err
		}
	}
	return nil
}

func (v *Volume) getField(field reflect.StructField, c *config.Config, iniOverrides *ini.File, structTag reflect.StructTag) (value interface{}) {
	//log.Debugf("Getting field %s from docker label", field.Name)
	var label string
	if prefix := structTag.Get("label"); prefix != "" {
		label = fmt.Sprintf("%s.", prefix)
	}
	label += field.Tag.Get("label")
	value, _ = v.getVolumeLabel(label)
	if value != "" {
		return
	}

	//log.Debugf("Getting field %s from ini overrides", field.Name)
	value = getIniValue(iniOverrides, structTag.Get("ini"), field.Tag.Get("ini"))
	if value != "" {
		return
	}

	//log.Debugf("Getting field %s from general config", field.Name)
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

	//log.Debugf("[%s] field: %s, value:%s", fieldType.Name, field.Kind(), v.Kind())

	if field.Kind() == v.Kind() {
		//log.Debugf("field and value have the same kind: %v", field.Kind())
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

func (v *Volume) getVolumeLabel(key string) (value string, err error) {
	//log.Debugf("Getting value for label %s of volume %s", key, vol.Name)
	value, ok := v.Labels[v.LabelPrefix+key]
	if !ok {
		errMsg := fmt.Sprintf("Key %v not found in labels for volume %v", key, v.Name)
		err = errors.New(errMsg)
	}
	return
}
