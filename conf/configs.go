package conf

import (
	"fmt"
	"html/template"
	"os"
	"reflect"
	"strconv"
	"strings"

	"os/signal"
	"sync"
	"syscall"

	"github.com/alecthomas/log4go"
)

var (
	cfg        *YamlCfg
	configLock = new(sync.RWMutex)
)

//热加载配置文件
func init() {
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT) //syscall.SIGUSR1
	go func() {
		for {
			<-s
			log4go.Info("Reloaded config.")
			Load("config.yaml")
		}
	}()
}

//cfgFile filepath
func Load(cfgFile string) error {

	var err error

	cfg, err = yamlReader(cfgFile)
	if err != nil {
		log4go.Error("Read Yaml File Failed: %s.", err.Error())
		return err
	}

	return nil
}

func getCfg() *YamlCfg {
	if cfg == nil {
		log4go.Debug("*YamlCfg is not initialized. Reinitializing....")
		cfg = new(YamlCfg)
	}
	return cfg
}

func String(path string) string {
	value, err := getCfg().String(path)

	if err == nil {
		return value
	}

	log4go.Error("fetch string value failed: %s", err.Error())

	return ""
}

func StringList(path string) []string {
	value, err := getCfg().List(path)

	if err == nil {

		return StrList(value)
	}

	log4go.Error("fetch StringList value failed: %s", err.Error())

	return nil
}

func Int(path string) int {
	value, err := getCfg().Int(path)

	if err == nil {
		return value
	}

	log4go.Error("fetch integer value failed: %s", err.Error())

	return 0
}

func Float64(path string) float64 {
	value, err := getCfg().Float64(path)

	if err == nil {
		return value
	}

	log4go.Error("fetch float value failed: %s", err.Error())

	return float64(0)
}

func Bool(path string) bool {
	value, err := getCfg().Bool(path)

	if err == nil {
		return value
	}

	log4go.Error("fetch bool value failed: %s", err.Error())

	return false
}

func strList(i interface{}) ([]string, error) {
	var a []string

	switch v := i.(type) {
	case []interface{}:
		for _, u := range v {
			a = append(a, ToString(u))
		}
		return a, nil
	case []string:
		return v, nil
	case string:
		return strings.Fields(v), nil
	case interface{}:
		str, err := toString(v)
		if err != nil {
			return a, fmt.Errorf("unable to cast %#v of type %T to []string", i, i)
		}
		return []string{str}, nil
	default:
		return a, fmt.Errorf("unable to cast %#v of type %T to []string", i, i)
	}
}

func StrList(i interface{}) []string {
	v, err := strList(i)
	if err != nil {
		log4go.Error(err.Error())
	}
	return v
}

func indirectToStringerOrError(a interface{}) interface{} {
	if a == nil {
		return nil
	}

	var errorType = reflect.TypeOf((*error)(nil)).Elem()
	var fmtStringerType = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()

	v := reflect.ValueOf(a)
	for !v.Type().Implements(fmtStringerType) && !v.Type().Implements(errorType) && v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	return v.Interface()
}

func ToString(i interface{}) string {
	v, err := toString(i)
	if err != nil {
		log4go.Error(err.Error())
	}
	return v
}

func toString(i interface{}) (string, error) {
	i = indirectToStringerOrError(i)

	switch s := i.(type) {
	case string:
		return s, nil
	case bool:
		return strconv.FormatBool(s), nil
	case float64:
		return strconv.FormatFloat(s, 'f', -1, 64), nil
	case float32:
		return strconv.FormatFloat(float64(s), 'f', -1, 32), nil
	case int:
		return strconv.Itoa(s), nil
	case int64:
		return strconv.FormatInt(s, 10), nil
	case int32:
		return strconv.Itoa(int(s)), nil
	case int16:
		return strconv.FormatInt(int64(s), 10), nil
	case int8:
		return strconv.FormatInt(int64(s), 10), nil
	case uint:
		return strconv.FormatInt(int64(s), 10), nil
	case uint64:
		return strconv.FormatInt(int64(s), 10), nil
	case uint32:
		return strconv.FormatInt(int64(s), 10), nil
	case uint16:
		return strconv.FormatInt(int64(s), 10), nil
	case uint8:
		return strconv.FormatInt(int64(s), 10), nil
	case []byte:
		return string(s), nil
	case template.HTML:
		return string(s), nil
	case template.URL:
		return string(s), nil
	case template.JS:
		return string(s), nil
	case template.CSS:
		return string(s), nil
	case template.HTMLAttr:
		return string(s), nil
	case nil:
		return "", nil
	case fmt.Stringer:
		return s.String(), nil
	case error:
		return s.Error(), nil
	default:
		return "", fmt.Errorf("unable to cast %#v of type %T to string", i, i)
	}
}

func LogEnv(path string) bool {
	if _, err := os.Stat(path); err != nil {
		log4go.Warn(err.Error())
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
