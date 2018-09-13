// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package x509

import (
	"encoding/pem"
	"errors"
	"runtime"
	"strings"
	"testing"
	"time"
	"io/ioutil"
	"log"
	"os"
	"crypto/rand"
	"bufio"
	"crypto"
	"crypto/rsa"
	"encoding/base64"

	"github.com/google/certificate-transparency-go/x509/pkix"
)

var supportSHA2 = true

type verifyTest struct {
	leaf                           string
	leafKey                        string
	intermediates                  []string
	roots                          []string
	currentTime                    int64
	dnsName                        string
	systemSkip                     bool
	keyUsages                      []ExtKeyUsage
	testSystemRootsError           bool
	sha2                           bool
	disableTimeChecks              bool
	disableCriticalExtensionChecks bool
	disableNameChecks              bool

	errorCallback  func(*testing.T, int, error) bool
	expectedChains [][]string
}

var verifyTests = []verifyTest{
	{
		leaf:          "/home/tang/cert/chain/ExampleServer.crt",
		leafKey:       "/home/tang/cert/chain/ExampleServer.key",
		intermediates: []string{"/home/tang/cert/chain/ExampleIntermediateCA.crt"},
		roots:         []string{"/home/tang/cert/chain/ExampleRootCA.crt"},
		currentTime:   1536739579,
		dnsName:       "*.example.com",

		expectedChains: [][]string{
			{
				"*.example.com",
				"Example Intermediate CA",
				"Example Root CA",
			},
		},
	},
}

func readPEM(pemName string) []byte {
	pemBytes, err := ioutil.ReadFile(pemName)
	if err != nil {
		panic(err)
	}

	return pemBytes
}

func certificateFromPEM(pemName string) (*Certificate, error) {
	pemBytes := readPEM(pemName)
	block, _ := pem.Decode([]byte(pemBytes))
	if block == nil {
		return nil, errors.New("failed to decode PEM")
	}
	return ParseCertificate(block.Bytes)
}

func testVerify(t *testing.T, useSystemRoots bool) {
	for i, test := range verifyTests {
		if useSystemRoots && test.systemSkip {
			continue
		}
		if runtime.GOOS == "windows" && test.testSystemRootsError {
			continue
		}
		if useSystemRoots && !supportSHA2 && test.sha2 {
			continue
		}

		opts := VerifyOptions{
			Intermediates:                  NewCertPool(),
			DNSName:                        test.dnsName,
			CurrentTime:                    time.Unix(test.currentTime, 0),
			KeyUsages:                      test.keyUsages,
			DisableTimeChecks:              test.disableTimeChecks,
			DisableCriticalExtensionChecks: test.disableCriticalExtensionChecks,
			DisableNameChecks:              test.disableNameChecks,
		}

		if !useSystemRoots {
			opts.Roots = NewCertPool()
			for j, root := range test.roots {
				ok := opts.Roots.AppendCertsFromPEM(readPEM(root))
				if !ok {
					t.Errorf("#%d: failed to parse root #%d", i, j)
					return
				}
			}
		}

		for j, intermediate := range test.intermediates {
			ok := opts.Intermediates.AppendCertsFromPEM(readPEM(intermediate))
			if !ok {
				t.Errorf("#%d: failed to parse intermediate #%d", i, j)
				return
			}
		}

		leaf, err := certificateFromPEM(test.leaf)
		if IsFatal(err) {
			t.Errorf("#%d: failed to parse leaf: %s", i, err)
			return
		}

		var oldSystemRoots *CertPool
		if test.testSystemRootsError {
			oldSystemRoots = systemRootsPool()
			systemRoots = nil
			opts.Roots = nil
		}

		chains, err := leaf.Verify(opts)

		if test.testSystemRootsError {
			systemRoots = oldSystemRoots
		}

		if test.errorCallback == nil && err != nil {
			t.Errorf("#%d: unexpected error: %s", i, err)
		}
		if test.errorCallback != nil {
			if !test.errorCallback(t, i, err) {
				return
			}
		}

		if len(chains) != len(test.expectedChains) {
			t.Errorf("#%d: wanted %d chains, got %d", i, len(test.expectedChains), len(chains))
		}

		// We check that each returned chain matches a chain from
		// expectedChains but an entry in expectedChains can't match
		// two chains.
		seenChains := make([]bool, len(chains))
	NextOutputChain:
		for _, chain := range chains {
		TryNextExpected:
			for j, expectedChain := range test.expectedChains {
				if seenChains[j] {
					continue
				}
				if len(chain) != len(expectedChain) {
					continue
				}
				for k, cert := range chain {
					if !strings.Contains(nameToKey(&cert.Subject), expectedChain[k]) {
						continue TryNextExpected
					}
				}
				// we matched
				seenChains[j] = true
				continue NextOutputChain
			}
			t.Errorf("#%d: No expected chain matched %s", i, chainToDebugString(chain))
		}

		message := "Hello World"
		privateKeyStr, err := readPrivateKey(test.leafKey)
		if err != nil {
			log.Fatal(err)
		}

		signature, err := createSignature(message, privateKeyStr)
		if err != nil {
			log.Fatal(err)
		}

		if err := verifySignature(leaf, message, signature); err != nil {
			log.Fatal("err: ", err)
		}
	}
}

func verifySignature(cert *Certificate, message string, signature string) error {
    signDataByte, err := base64.StdEncoding.DecodeString(signature)
    if err != nil {
        return err
    }

    h := crypto.Hash.New(crypto.SHA256)
    h.Write([]byte(message))
    hashed := h.Sum(nil)

    err = rsa.VerifyPKCS1v15(cert.PublicKey.(*rsa.PublicKey), crypto.SHA256, hashed, signDataByte)
    if err != nil {
        return err
    }

    return nil
}

func readPrivateKey(filepath string) (string, error) {
    s := ""
    fp, err := os.Open(filepath)
    if err != nil {
        return "", err
    }
    defer fp.Close()
    scanner := bufio.NewScanner(fp)
    for scanner.Scan() {
        text := scanner.Text()
        if text == "-----BEGIN RSA PRIVATE KEY-----" || text == "-----END RSA PRIVATE KEY-----" {
            continue
        }
        s = s + scanner.Text()
    }
    if err := scanner.Err(); err != nil {
        return "", err
    }

    return s, nil
}

func createSignature(message, keystr string) (string, error) {
    keyBytes, err := base64.StdEncoding.DecodeString(keystr)
    if err != nil {
        return "", err
    }

    private, err := ParsePKCS1PrivateKey(keyBytes)
    if err != nil {
        return "", err
    }

    h := crypto.Hash.New(crypto.SHA256)
    h.Write(([]byte)(message))
    hashed := h.Sum(nil)

    signedData, err := rsa.SignPKCS1v15(rand.Reader, private, crypto.SHA256, hashed)
    if err != nil {
        return "", err
    }

    signature := base64.StdEncoding.EncodeToString(signedData)
    return signature, nil
}

func TestGoVerify(t *testing.T) {
	testVerify(t, false)
}

func chainToDebugString(chain []*Certificate) string {
	var chainStr string
	for _, cert := range chain {
		if len(chainStr) > 0 {
			chainStr += " -> "
		}
		chainStr += nameToKey(&cert.Subject)
	}
	return chainStr
}

func nameToKey(name *pkix.Name) string {
	return strings.Join(name.Country, ",") + "/" + strings.Join(name.Organization, ",") + "/" + strings.Join(name.OrganizationalUnit, ",") + "/" + name.CommonName
}
