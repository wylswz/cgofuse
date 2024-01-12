package args

import "strings"

const (
	argEps = "--datasource="
)

func ResolveDatasource(args []string) string {
	for _, arg := range args {
		if strings.Contains(arg, argEps) {
			return strings.Replace(arg, argEps, "", -1)
		}
	}
	return "etcd://127.0.0.1:2379"
}
