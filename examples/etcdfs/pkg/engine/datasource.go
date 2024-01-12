package engine

import "net/url"

type Impl string

const (
	EtcdImpl = "etcd"
)

type Datasource struct {
	url *url.URL
}

type Authentication struct {
	Username string
	Password string
}

func (d *Datasource) GetScheme() string {
	return d.url.Scheme
}

func (d *Datasource) GetAuth() *Authentication {
	var user *url.Userinfo
	if user = d.url.User; user != nil {
		auth := Authentication{
			Username: d.url.User.Username(),
		}
		if pswd, ok := user.Password(); ok {
			auth.Password = pswd
		}
		return &auth
	}
	return nil
}

func (d *Datasource) GetHost() string {
	return d.url.Hostname()
}

func (d *Datasource) GetPort() string {
	return d.url.Port()
}

func NewDatasource(connStr string) (*Datasource, error) {
	connUrl, err := url.Parse(connStr)
	if err != nil {
		return nil, ErrBadUrl
	}
	return &Datasource{
		url: connUrl,
	}, nil
}
