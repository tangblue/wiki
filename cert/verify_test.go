// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package certtest

import (
	"bufio"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/certificate-transparency-go/x509"
	"github.com/google/certificate-transparency-go/x509/pkix"
)

var supportSHA2 = true

type verifyTest struct {
	leaf              string
	leafKey           string
	intermediates     []string
	roots             []string
	currentTime       int64
	dnsName           string
	systemSkip        bool
	keyUsages         []x509.ExtKeyUsage
	sha2              bool
	disableTimeChecks bool

	errorCallback  func(*testing.T, int, error) bool
	expectedChains [][]string
}

var verifyTests = []verifyTest{
	{
		leaf:          "ExampleServer.crt",
		leafKey:       "ExampleServer.key",
		intermediates: []string{"ExampleIntermediateCA.crt"},
		roots:         []string{"ExampleRootCA.crt"},
		currentTime:   0,
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

func certificateFromPEM(pemName string) (*x509.Certificate, error) {
	pemBytes := readPEM(pemName)
	block, _ := pem.Decode([]byte(pemBytes))
	if block == nil {
		return nil, errors.New("failed to decode PEM")
	}
	return x509.ParseCertificate(block.Bytes)
}

func testVerify(t *testing.T, useSystemRoots bool) {
	for i, test := range verifyTests {
		if useSystemRoots && test.systemSkip {
			continue
		}
		if useSystemRoots && !supportSHA2 && test.sha2 {
			continue
		}

		currentTime := test.currentTime
		if currentTime == 0 {
			currentTime = time.Now().Unix()
		}
		opts := x509.VerifyOptions{
			Intermediates:     x509.NewCertPool(),
			DNSName:           test.dnsName,
			CurrentTime:       time.Unix(currentTime, 0),
			KeyUsages:         test.keyUsages,
			DisableTimeChecks: test.disableTimeChecks,
		}

		if !useSystemRoots {
			opts.Roots = x509.NewCertPool()
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
		if err != nil {
			t.Errorf("#%d: failed to parse leaf: %s", i, err)
			return
		}

		chains, err := leaf.Verify(opts)

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

func verifySignature(cert *x509.Certificate, message string, signature string) error {
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

	private, err := x509.ParsePKCS1PrivateKey(keyBytes)
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

func chainToDebugString(chain []*x509.Certificate) string {
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
