//用来缓存OCSP文件，因为有些时候请求的校验服务器（如https://ocsp.int-x3.letsencrypt.org）被墙了，
//在本机获取OCSP的校验信息保存下来，让服务器直接读取使用

package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"golang.org/x/crypto/ocsp"
	"io/ioutil"
	"minGateway/util"
	"os"
)

var (
	fileDir = "/Users/chunyongchen/GoProjects/minGateway/bin/cert/"
	pemName = "truth.juleu.com.pem"
	keyName = "truth.juleu.com.key"
)

func main() {
	getOcspCache()
}

func getOcspCache() {
	certFile := fileDir + pemName
	keyFile := fileDir + keyName
	crt, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(certFile)
	fmt.Println(keyFile)

	OCSPBuf, _, certx509, err := GetOCSPForCert(crt.Certificate)
	if err == nil {
		//写文件
		for _, v := range certx509.DNSNames {
			name := v + ".ocsp"
			ocspFile := fileDir + name
			if util.PathExists(ocspFile) {
				_ = os.Remove(ocspFile)
			}
			_ = ioutil.WriteFile(ocspFile, OCSPBuf, 0644)
			fmt.Println(name + " file write success")
		}

	} else {
		fmt.Println(err)
	}
}

func GetOCSPForCert(cert [][]byte) ([]byte, *ocsp.Response, *x509.Certificate, error) {

	bundle := new(bytes.Buffer)
	for _, derBytes := range cert {
		err := pem.Encode(bundle, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
		if err != nil {
			fmt.Println(err)
			return nil, nil, nil, err
		}
	}
	pemBundle := bundle.Bytes()

	certificates, err := parsePEMBundle(pemBundle)
	if err != nil {
		return nil, nil, nil, err
	}

	// We expect the certificate slice to be ordered downwards the chain.
	// SRV CRT -> CA. We need to pull the leaf and issuer certs out of it,
	// which should always be the first two certificates. If there's no
	// OCSP server listed in the leaf cert, there's nothing to do. And if
	// we have only one certificate so far, we need to get the issuer cert.
	issuedCert := certificates[0]
	if len(issuedCert.OCSPServer) == 0 {
		return nil, nil, nil, errors.New("no OCSP server specified in cert")
	}
	if len(certificates) == 1 {
		// TODO: build fallback. If this fails, check the remaining array entries.
		if len(issuedCert.IssuingCertificateURL) == 0 {
			return nil, nil, issuedCert, errors.New("no issuing certificate URL")
		}

		resp, errC := util.HttpGet(issuedCert.IssuingCertificateURL[0])
		if errC != nil {
			return nil, nil, issuedCert, errC
		}
		defer resp.Body.Close()

		issuerBytes, errC := ioutil.ReadAll(util.LimitReader(resp.Body, 1024*1024))
		if errC != nil {
			return nil, nil, issuedCert, errC
		}

		issuerCert, errC := x509.ParseCertificate(issuerBytes)
		if errC != nil {
			return nil, nil, issuedCert, errC
		}

		// Insert it into the slice on position 0
		// We want it ordered right SRV CRT -> CA
		certificates = append(certificates, issuerCert)
	}
	issuerCert := certificates[1]

	// Finally kick off the OCSP request.
	ocspReq, err := ocsp.CreateRequest(issuedCert, issuerCert, nil)
	if err != nil {
		return nil, nil, issuedCert, err
	}

	reader := bytes.NewReader(ocspReq)
	fmt.Println("ocsp server url:", issuedCert.OCSPServer[0])
	req, err := util.HttpPost(issuedCert.OCSPServer[0], "application/ocsp-request", reader)
	if err != nil {
		return nil, nil, issuedCert, err
	}
	defer req.Body.Close()

	ocspResBytes, err := ioutil.ReadAll(util.LimitReader(req.Body, 1024*1024))
	if err != nil {
		return nil, nil, issuedCert, err
	}

	ocspRes, err := ocsp.ParseResponse(ocspResBytes, issuerCert)
	if err != nil {
		return nil, nil, issuedCert, err
	}

	return ocspResBytes, ocspRes, issuedCert, nil
}

// parsePEMBundle parses a certificate bundle from top to bottom and returns
// a slice of x509 certificates. This function will error if no certificates are found.
func parsePEMBundle(bundle []byte) ([]*x509.Certificate, error) {
	var certificates []*x509.Certificate
	var certDERBlock *pem.Block

	for {
		certDERBlock, bundle = pem.Decode(bundle)
		if certDERBlock == nil {
			break
		}

		if certDERBlock.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(certDERBlock.Bytes)
			if err != nil {
				return nil, err
			}
			certificates = append(certificates, cert)
		}
	}

	if len(certificates) == 0 {
		return nil, errors.New("no certificates were found while parsing the bundle")
	}

	return certificates, nil
}
