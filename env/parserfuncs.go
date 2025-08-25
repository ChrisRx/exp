package env

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"net/url"
	"reflect"
	"time"

	"go.chrisrx.dev/x/ptr"
)

type CustomParserFunc func(Field, string) (any, error)

var customParserFuncs = make(map[reflect.Type]CustomParserFunc)

func init() {
	Register[time.Time](func(field Field, s string) (any, error) {
		return time.Parse(field.Layout, s)
	})

	Register[time.Duration](func(field Field, s string) (any, error) {
		return time.ParseDuration(s)
	})

	Register[url.URL](func(field Field, s string) (any, error) {
		u, err := url.Parse(s)
		if err != nil {
			return nil, err
		}
		return ptr.From(u), nil
	})

	Register[[]byte](func(field Field, s string) (any, error) {
		return []byte(s), nil
	})

	Register[net.IP](func(field Field, s string) (any, error) {
		return net.ParseIP(s), nil
	})

	Register[rsa.PublicKey](func(field Field, s string) (any, error) {
		pub, err := loadPublicKey([]byte(s))
		if err != nil {
			return nil, err
		}
		if key, ok := pub.(*rsa.PublicKey); ok {
			return ptr.From(key), nil
		}
		return nil, fmt.Errorf("expected *rsa.PublicKey, received %T", pub)
	})

	Register[x509.Certificate](func(field Field, s string) (any, error) {
		cert, err := loadPublicKey([]byte(s))
		if err != nil {
			return nil, err
		}
		if key, ok := cert.(*x509.Certificate); ok {
			return ptr.From(key), nil
		}
		return nil, fmt.Errorf("expected *x509.Certificate, received %T", cert)
	})
}

func Register[T any](fn CustomParserFunc) {
	rt := reflect.TypeFor[T]()
	if rt.Kind() == reflect.Pointer {
		panic(fmt.Errorf("cannot register type %v: must not be pointer", rt))
	}
	customParserFuncs[rt] = func(field Field, s string) (any, error) {
		// avoid needing customer parsers to handle empty input
		if s == "" {
			return nil, nil
		}
		return fn(field, s)
	}
}

// TODO(ChrisRx): move to another package
func loadPublicKey(data []byte) (any, error) {
	block, _ := pem.Decode(data)
	if block != nil {
		data = block.Bytes
	}
	if pub, err := x509.ParsePKIXPublicKey(data); err == nil {
		return pub, nil
	}
	if cert, err := x509.ParseCertificate(data); err == nil {
		return cert.PublicKey, nil
	}
	return nil, fmt.Errorf("cannot parse public key")
}
