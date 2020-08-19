/*
OCSP（Online Certificate Status Protocol，在线证书状态协议）是用来检验证书合法性的在线查询服务，一般由证书所属 CA 提供。某些客户端会在 TLS 握手阶段
进一步协商时，实时查询 OCSP 接口，并在获得结果前阻塞后续流程。OCSP 查询本质是一次完整的 HTTP 请求 - 响应，这中间 DNS 查询、建立 TCP、服务端处理等环节
都可能耗费很长时间，导致最终建立 TLS 连接时间变得更长。

OCSP Stapling（OCSP 封套），是指服务端主动获取 OCSP 查询结果并随着证书一起发送给客户端，从而让客户端跳过自己去验证的过程，提高 TLS 握手效率。

下面方法能在线获取OCSP信息，获取后将它设置到tls的OCSPStaple上，能开启go服务的OCSP Stapling

运行服务后，通过 openssl s_client -connect cspi.juleu.com:443 -status -tlsextdebug < /dev/null 2>&1 | grep -i "OCSP response" 查询是否开启
正常应该返回：
OCSP response:
OCSP Response Data:
    OCSP Response Status: successful (0x0)
    Response Type: Basic OCSP Response
*/

package main

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"golang.org/x/crypto/ocsp"
	"io/ioutil"
	"minGateway/util"
)

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
