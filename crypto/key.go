package crypto

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"
)

type RemotePublicKey struct {
	url          string
	publicKey    *rsa.PublicKey
	lastUpdated  time.Time
	refreshEvery time.Duration
	mu           sync.RWMutex
}

func NewRemotePublicKey(url string, refreshEvery time.Duration) (*RemotePublicKey, error) {
	r := &RemotePublicKey{
		url:          url,
		refreshEvery: refreshEvery,
	}
	if err := r.refresh(); err != nil {
		return nil, err
	}
	go r.autoRefresh()
	return r, nil
}

func (r *RemotePublicKey) autoRefresh() {
	ticker := time.NewTicker(r.refreshEvery)
	for range ticker.C {
		_ = r.refresh()
	}
}

func (r *RemotePublicKey) refresh() error {
	resp, err := http.Get(r.url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	block, _ := pem.Decode(raw)
	if block == nil {
		return errors.New("invalid PEM")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return errors.New("not RSA public key")
	}

	r.mu.Lock()
	r.publicKey = rsaPub
	r.lastUpdated = time.Now()
	r.mu.Unlock()

	return nil
}

func (r *RemotePublicKey) Get() *rsa.PublicKey {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.publicKey
}
