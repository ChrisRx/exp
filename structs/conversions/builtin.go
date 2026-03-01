package conversions

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"net/url"
	"time"

	"go.chrisrx.dev/x/convert"
)

func init() {
	convert.Register(func(s string, opts ...convert.Option) (time.Time, error) {
		o := convert.NewOptions(opts)
		return time.Parse(o.Layout, s)
	})
	convert.Register(func(t time.Time, opts ...convert.Option) (string, error) {
		o := convert.NewOptions(opts)
		return t.Format(o.Layout), nil
	})

	convert.Register(func(s string, opts ...convert.Option) (time.Duration, error) {
		return time.ParseDuration(s)
	})

	convert.Register(func(s string, opts ...convert.Option) (*url.URL, error) {
		return url.Parse(s)
	})

	convert.Register(func(s string, opts ...convert.Option) ([]byte, error) {
		return []byte(s), nil
	})

	convert.Register(func(s string, opts ...convert.Option) (net.HardwareAddr, error) {
		return net.HardwareAddr(s), nil
	})

	convert.Register(func(s string, opts ...convert.Option) (net.IP, error) {
		return net.ParseIP(s), nil
	})

	convert.Register(func(s string, opts ...convert.Option) (*rsa.PublicKey, error) {
		pub, err := loadPublicKey([]byte(s))
		if err != nil {
			return nil, err
		}
		if key, ok := pub.(*rsa.PublicKey); ok {
			return key, nil
		}
		return nil, fmt.Errorf("expected *rsa.PublicKey, received %T", pub)
	})

	convert.Register(func(s string, opts ...convert.Option) (*x509.Certificate, error) {
		cert, err := loadPublicKey([]byte(s))
		if err != nil {
			return nil, err
		}
		if key, ok := cert.(*x509.Certificate); ok {
			return key, nil
		}
		return nil, fmt.Errorf("expected *x509.Certificate, received %T", cert)
	})
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
