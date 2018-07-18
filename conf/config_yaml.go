package conf

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/go-yaml/yaml"
)

type YamlCfg struct {
	Yaml interface{}
}

func yamlReader(filename string) (*YamlCfg, error) {
	cfg, err := ioutil.ReadFile(filename) //ReadFile 从filename指定的文件中读取数据并返回文件的内容.[]byte类型
	if err != nil {
		return nil, err
	}
	return unmarshalYaml(cfg)
}

func unmarshalYaml(cfg []byte) (*YamlCfg, error) {

	var out interface{}
	var err error

	if err = yaml.Unmarshal(cfg, &out); err != nil {
		return nil, err
	}
	if out, err = normalizeValue(out); err != nil {
		return nil, err
	}

	return &YamlCfg{Yaml: out}, nil
}

func normalizeValue(value interface{}) (interface{}, error) {
	switch value := value.(type) {

	case map[interface{}]interface{}:
		//fmt.Println("map[interface{}]interface{}:\n", value)
		node := make(map[string]interface{}, len(value))
		for k, v := range value {

			key, ok := k.(string)
			if !ok {
				return nil, fmt.Errorf("Unsupported map key: %#v", k)
			}
			item, err := normalizeValue(v)
			if err != nil {
				return nil, fmt.Errorf("Unsupported map value: %#v", v)
			}
			node[key] = item
		}
		return node, nil
	case map[string]interface{}:
		//fmt.Println("map[string]interface{}:\n", value)
		node := make(map[string]interface{}, len(value))
		for key, v := range value {

			item, err := normalizeValue(v)
			if err != nil {
				return nil, fmt.Errorf("Unsupported map value: %#v", v)
			}
			node[key] = item
		}
		return node, nil
	case []interface{}:
		//fmt.Println("[]interface{}:\n", value)
		node := make([]interface{}, len(value))
		for key, v := range value {

			item, err := normalizeValue(v)
			if err != nil {
				return nil, fmt.Errorf("Unsupported list item: %#v", v)
			}
			node[key] = item
		}
		return node, nil
	case bool, float64, int, string, nil:
		return value, nil
	}
	return nil, fmt.Errorf("Unsupported type: %T", value)
}

//get cfg
func Get(cfg interface{}, path string) (interface{}, error) {
	parts := strings.Split(path, ".")
	// Normalize path.
	for k, v := range parts {
		if v == "" {
			if k == 0 {
				parts = parts[1:]
			} else {
				return nil, fmt.Errorf("Invalid path %q", path)
			}
		}
	}
	//fmt.Println("parts:", parts)
	// Get the value.
	for pos, part := range parts {
		switch c := cfg.(type) {
		case []interface{}:
			//fmt.Println("[]interface{}:", cfg)
			if i, err := strconv.ParseInt(part, 10, 0); err == nil {
				if int(i) < len(c) {
					cfg = c[i]
				} else {
					return nil, fmt.Errorf(
						"Index out of range at %q: list has only %v items",
						strings.Join(parts[:pos+1], "."), len(c))
				}
			} else {
				return nil, fmt.Errorf("Invalid list index at %q",
					strings.Join(parts[:pos+1], "."))
			}
		case map[string]interface{}:
			//fmt.Println("map[string]interface{}:", cfg)
			if value, ok := c[part]; ok {
				cfg = value
			} else {
				return nil, fmt.Errorf("Nonexistent map key at %q",
					strings.Join(parts[:pos+1], "."))
			}
		default:
			//fmt.Println("default:", cfg)
			return nil, fmt.Errorf(
				"Invalid type at %q: expected []interface{} or map[string]interface{}; got %T",
				strings.Join(parts[:pos+1], "."), cfg)
		}
	}

	return cfg, nil
}

func (cfg *YamlCfg) Bool(path string) (bool, error) {
	n, err := Get(cfg.Yaml, path)
	if err != nil {
		return false, err
	}
	switch n := n.(type) {
	case bool:
		return n, nil
	case string:
		return strconv.ParseBool(n)
	}
	return false, typeException("bool or string", n)
}

func (cfg *YamlCfg) Float64(path string) (float64, error) {
	n, err := Get(cfg.Yaml, path)
	if err != nil {
		return 0, err
	}
	switch n := n.(type) {
	case float64:
		return n, nil
	case int:
		return float64(n), nil
	case string:
		return strconv.ParseFloat(n, 64)
	}
	return 0, typeException("float64, int or string", n)
}

func (cfg *YamlCfg) Int(path string) (int, error) {
	n, err := Get(cfg.Yaml, path)
	if err != nil {
		return 0, err
	}
	switch n := n.(type) {
	case float64:
		if i := int(n); fmt.Sprint(i) == fmt.Sprint(n) {
			return i, nil
		} else {
			return 0, fmt.Errorf("Value can't be converted to int: %v", n)
		}
	case int:
		return n, nil
	case string:
		if v, err := strconv.ParseInt(n, 10, 0); err == nil {
			return int(v), nil
		} else {
			return 0, err
		}
	}
	return 0, typeException("float64, int or string", n)
}

func (cfg *YamlCfg) List(path string) ([]interface{}, error) {
	n, err := Get(cfg.Yaml, path)
	if err != nil {
		return nil, err
	}
	if value, ok := n.([]interface{}); ok {
		return value, nil
	}
	return nil, typeException("[]interface{}", n)
}

func (cfg *YamlCfg) String(path string) (string, error) {
	n, err := Get(cfg.Yaml, path)
	if err != nil {
		return "", err
	}
	switch n := n.(type) {
	case bool, float64, int:
		return fmt.Sprint(n), nil
	case string:
		return n, nil
	}
	return "", typeException("bool, float64, int or string", n)
}

func typeException(wannaType string, gotType interface{}) error {
	return fmt.Errorf("Type exception: you wanna %s, but, regrettably, got %T", wannaType, gotType)
}
