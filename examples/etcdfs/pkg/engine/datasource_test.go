package engine

import (
	"reflect"
	"testing"
)

func TestEtcdDataSource(t *testing.T) {
	testCases := []struct {
		connStr string

		err    error
		auth   *Authentication
		host   string
		port   string
		scheme string
	}{
		{
			connStr: "etcd://user:pass@127.0.0.1,127.0.0.2:2379",

			err:    nil,
			auth:   &Authentication{"user", "pass"},
			host:   "127.0.0.1,127.0.0.2",
			port:   "2379",
			scheme: "etcd",
		},
	}

	for _, testCase := range testCases {
		ds, err := NewDatasource(testCase.connStr)
		if testCase.err == nil {
			assertNil(t, err)
		}
		assertEq(t, ds.GetAuth(), testCase.auth)
		assertEq(t, ds.GetPort(), testCase.port)
		assertEq(t, ds.GetHost(), testCase.host)
		assertEq(t, ds.GetScheme(), testCase.scheme)
	}
}

func assertNil(t *testing.T, v interface{}) {
	if v != nil {
		t.Fail()
	}
}

func assertEq(t *testing.T, a, b interface{}) {
	if !reflect.DeepEqual(a, b) {
		t.Fail()
	}
}
