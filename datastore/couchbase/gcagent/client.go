//  Copyright 2020-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included in
//  the file licenses/Couchbase-BSL.txt.  As of the Change Date specified in that
//  file, in accordance with the Business Source License, use of this software will
//  be governed by the Apache License, Version 2.0, included in the file
//  licenses/APL.txt.

package gcagent

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"github.com/couchbase/cbauth"
	gctx "github.com/couchbase/gocbcore-transactions"
	"github.com/couchbase/gocbcore/v9"
	"github.com/couchbase/gocbcore/v9/connstr"
	ntls "github.com/couchbase/goutils/tls"
)

const (
	_CONNECTTIMEOUT   = 10000 * time.Millisecond
	_KVCONNECTTIMEOUT = 7000 * time.Millisecond
	_KVTIMEOUT        = 2500 * time.Millisecond
	_WARMUPTIMEOUT    = 1000 * time.Millisecond
	_WARMUP           = false
	_CLOSEWAIT        = 2 * time.Minute
	_kVPOOLSIZE       = 8
	_MAXQUEUESIZE     = 32 * 1024
)

type MemcachedAuthProvider struct {
	c *Client
}

func (auth *MemcachedAuthProvider) Credentials(req gocbcore.AuthCredsRequest) (
	[]gocbcore.UserPassPair, error) {
	endpoint := req.Endpoint

	// get rid of the http:// or https:// prefix from the endpoint
	endpoint = strings.TrimPrefix(strings.TrimPrefix(endpoint, "http://"), "https://")
	username, password, err := cbauth.GetMemcachedServiceAuth(endpoint)
	if err != nil {
		return []gocbcore.UserPassPair{{}}, err
	}

	return []gocbcore.UserPassPair{{
		Username: username,
		Password: password,
	}}, nil
}

func (auth *MemcachedAuthProvider) SupportsNonTLS() bool {
	return true
}

func (auth *MemcachedAuthProvider) SupportsTLS() bool {
	return true
}

func (auth *MemcachedAuthProvider) Certificate(req gocbcore.AuthCertRequest) (*tls.Certificate, error) {
	return nil, nil
	// At present when we act as client we use Credentials, not certificates.
	// return auth.c.certs, nil
}

// Call this method with a TLS certificate file name to make communication
type Client struct {
	config       *gocbcore.AgentConfig
	transactions *gctx.Manager
	rootCAs      *x509.CertPool
	mutex        sync.RWMutex
	sslConfigFn  SSLConfigFn
	atrLocations map[string]gctx.LostATRLocation
	certs        *tls.Certificate
}

type SSLConfigFn func() (*gocbcore.AgentConfig, error)

func NewClient(url string, sslHostFn func() (string, string), caFile, certFile, keyFile string, passphrase []byte) (rv *Client, err error) {
	var connSpec *connstr.ConnSpec

	rv = &Client{}

	// network=default use internal (vs  alternative) addresses
	// http bootstrap is faster
	options := "?network=default&bootstrap_on=http"
	if rv.config, connSpec, err = agentConfig(url, options, rv); err != nil {
		return nil, err
	}

	rv.sslConfigFn = func() (*gocbcore.AgentConfig, error) {
		// create SSL agent config file
		sslHost, sslPort := sslHostFn()
		if len(connSpec.Addresses) > 0 {
			if sslHost == "" {
				sslHost = connSpec.Addresses[0].Host
			}
			surl := "couchbases://" + sslHost
			if sslPort != "" {
				// couchbases schema with custom port will not allowed http bootstrap.
				surl = "http://" + sslHost + ":" + sslPort
			}
			sslConfig, _, err1 := agentConfig(surl, options, rv)
			return sslConfig, err1
		}
		return nil, fmt.Errorf("no ssl address")
	}

	if certFile != "" || caFile != "" || keyFile != "" {
		if err = rv.InitTLS(caFile, certFile, keyFile, passphrase); err != nil {
			return nil, err
		}
	}

	// generic provider
	rv.atrLocations = make(map[string]gctx.LostATRLocation, 32)

	return rv, err
}

func agentConfig(url, options string, rv *Client) (config *gocbcore.AgentConfig, cspec *connstr.ConnSpec, err error) {
	config = &gocbcore.AgentConfig{
		ConnectTimeout:       _CONNECTTIMEOUT,
		KVConnectTimeout:     _KVCONNECTTIMEOUT,
		UseCollections:       true,
		KvPoolSize:           _kVPOOLSIZE,
		MaxQueueSize:         _MAXQUEUESIZE,
		Auth:                 &MemcachedAuthProvider{rv},
		DefaultRetryStrategy: gocbcore.NewBestEffortRetryStrategy(nil),
	}

	var connSpec connstr.ConnSpec
	if connSpec, err = connstr.Parse(url + options); err == nil {
		err = config.FromConnStr(connSpec.String())
	}

	return config, &connSpec, err
}

func (c *Client) InitTransactions(txConfig *gctx.Config) (err error) {
	c.AddAtrLocation(&txConfig.CustomATRLocation)
	txConfig.LostCleanupATRLocationProvider = func() (lostAtrLocations []gctx.LostATRLocation, cerr error) {
		c.mutex.RLock()
		defer c.mutex.RUnlock()
		lostAtrLocations = make([]gctx.LostATRLocation, 0, len(c.atrLocations))
		for _, atrl := range c.atrLocations {
			lostAtrLocations = append(lostAtrLocations, atrl)
		}
		return
	}

	c.transactions, err = gctx.Init(txConfig)
	return err
}

func (c *Client) AddAtrLocation(atrLocation *gctx.ATRLocation) (err error) {
	if atrLocation != nil && atrLocation.Agent != nil && atrLocation.Agent.BucketName() != "" {
		lostAtr := gctx.LostATRLocation{BucketName: atrLocation.Agent.BucketName(),
			ScopeName:      "_default",
			CollectionName: "_default"}

		if atrLocation.ScopeName != "" {
			lostAtr.ScopeName = atrLocation.ScopeName
		}
		if atrLocation.CollectionName != "" {
			lostAtr.CollectionName = atrLocation.CollectionName
		}
		s := lostAtr.BucketName + "." + lostAtr.ScopeName + "." + lostAtr.CollectionName
		c.mutex.Lock()
		defer c.mutex.Unlock()
		if _, ok := c.atrLocations[s]; !ok {
			c.atrLocations[s] = lostAtr
		}
	}
	return
}

func (c *Client) RemoveAtrLocation(bucketName string) (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for s, atrl := range c.atrLocations {
		if atrl.BucketName == bucketName {
			delete(c.atrLocations, s)
		}
	}
	return
}

func (c *Client) CreateAgentProvider(bucketName string) (*AgentProvider, error) {
	ap := &AgentProvider{client: c, bucketName: bucketName}
	err := ap.CreateOrRefreshAgent()
	return ap, err
}

func (c *Client) Close() {
	if c.transactions != nil {
		c.transactions.Close()
	}
	c.transactions = nil
	c.mutex.Lock()
	c.rootCAs = nil
	c.mutex.Unlock()
}

// with the KV engine encrypted.
func (c *Client) InitTLS(caFile, certFile, keyFile string, passphrase []byte) error {
	if len(caFile) > 0 {
		certFile = caFile
	}
	serverCert, err := ioutil.ReadFile(certFile)
	if err != nil {
		return err
	}
	CA_Pool := x509.NewCertPool()
	CA_Pool.AppendCertsFromPEM(serverCert)
	certs, err := ntls.LoadX509KeyPair(certFile, keyFile, passphrase)
	if err != nil {
		return err
	}
	c.mutex.Lock()
	// Set values for certs and passphrase
	c.certs = &certs
	c.rootCAs = CA_Pool
	c.mutex.Unlock()
	return nil
}

func (c *Client) ClearTLS() {
	c.mutex.Lock()
	c.rootCAs = nil
	c.certs = nil
	c.mutex.Unlock()
}

func (c *Client) TLSRootCAs() *x509.CertPool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.rootCAs
}

func (c *Client) Transactions() *gctx.Manager {
	return c.transactions
}
