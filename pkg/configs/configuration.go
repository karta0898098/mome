package configs

import (
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/karta0898098/mome/pkg/logging"
)

// Configuration are contain all app config
type Configuration struct {
	Log  logging.Config `mapstructure:"log"`
	GRPC GRPCServer     `mapstructure:"grpc"`
}

func ExpandEnvHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	replaceENV := func(stringData, tmp string) string {
		if strings.HasPrefix(tmp, "${") && strings.HasSuffix(tmp, "}") {
			envVarValue := os.Getenv(strings.TrimPrefix(strings.TrimSuffix(tmp, "}"), "${"))
			if len(envVarValue) > 0 {
				return strings.ReplaceAll(stringData, tmp, envVarValue)
			}
		}
		return stringData
	}

	if f.Kind() == reflect.String {
		stringData := data.(string)
		re := regexp.MustCompile(`\${([^}]+)}`)
		if re.MatchString(stringData) {
			for _, m := range re.FindAllStringSubmatch(stringData, -1) {
				for _, mm := range m {
					stringData = replaceENV(stringData, mm)
				}
			}
			return stringData, nil
		}
	}
	return data, nil
}
