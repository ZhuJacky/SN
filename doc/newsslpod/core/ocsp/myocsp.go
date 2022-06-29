package ocsp

import (
	"bytes"
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"io/ioutil"
	"net/http"
	// "ocsp"
)

//GetOCSPdata  获取OCSP数据，ca 可以为nil，从终端证书中可以得到：颁发者密钥标识、颁发者名字HASH
func GetOCSPdata(ctx context.Context, cert, ca *x509.Certificate) *[]byte {
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	var (
		request []byte
		err     error
	)

	if ca == nil {
		request, err = createRequestNotIssuer(cert, nil)
	} else {
		request, err = CreateRequest(cert, ca, nil)
	}

	if err != nil {
		//fmt.Println("生成包出错")
		return nil
	}
	//fmt.Printf("%X\n", request)
	var Resp *http.Response
	cli := &http.Client{}
	for i := 0; i < 2; i++ {
		for _, ocspurl := range cert.OCSPServer {
			req, err := http.NewRequest(http.MethodPost, ocspurl, bytes.NewReader(request))
			if err != nil {
				fmt.Println("myocsp req err: " + err.Error())
				return nil
			}
			req = req.WithContext(ctx)
			req.Header.Set("Content-Type", "application/ocsp-request")

			Resp, err = cli.Do(req)
			if err == nil {

				defer Resp.Body.Close()
				data, _ := ioutil.ReadAll(Resp.Body)

				if len(data) < 6 {

					return nil
				}
				return &data

			}

		}
	}
	return nil

}

// opts is nil then sensible defaults are used.
func createRequestNotIssuer(cert *x509.Certificate, opts *RequestOptions) ([]byte, error) {
	hashFunc := opts.hash()

	// OCSP seems to be the only place where these raw hash identifiers are
	// used. I took the following from
	// http://msdn.microsoft.com/en-us/library/ff635603.aspx
	var hashOID asn1.ObjectIdentifier
	hashOID, ok := hashOIDs[hashFunc]
	if !ok {
		return nil, x509.ErrUnsupportedAlgorithm
	}

	if !hashFunc.Available() {
		return nil, x509.ErrUnsupportedAlgorithm
	}
	h := opts.hash().New()

	//issuerKeyHash := h.Sum(nil)

	issuerKeyHash := cert.AuthorityKeyId

	h.Write(cert.RawIssuer)
	//h.Write(issuer.RawSubject)
	issuerNameHash := h.Sum(nil)

	return asn1.Marshal(ocspRequest{
		tbsRequest{
			Version: 0,
			RequestList: []request{
				{
					Cert: certID{
						pkix.AlgorithmIdentifier{
							Algorithm:  hashOID,
							Parameters: asn1.RawValue{Tag: 5 /* ASN.1 NULL */},
						},
						issuerNameHash,
						issuerKeyHash,
						cert.SerialNumber,
					},
				},
			},
		},
	})
}
