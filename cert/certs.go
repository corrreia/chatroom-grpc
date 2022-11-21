package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"
)

func GenerateCACertKey(path string) error {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
    if err != nil {
        return err
    }
    keyBytes := x509.MarshalPKCS1PrivateKey(key)
    // PEM encoding of private key
    CAkeyPem := pem.EncodeToMemory(
        &pem.Block{
            Type:  "RSA PRIVATE KEY",
            Bytes: keyBytes,
        },
    )
    fmt.Println(string(CAkeyPem))
    
    notBefore := time.Now()
    notAfter := notBefore.Add(365*24*10*time.Hour)

    //Create certificate templet
    template := x509.Certificate{
        SerialNumber:          big.NewInt(0),
        Subject:               pkix.Name{CommonName: "CA"},
        SignatureAlgorithm:    x509.SHA256WithRSA,
        NotBefore:             notBefore,
        NotAfter:              notAfter,
        BasicConstraintsValid: true,
        KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyAgreement | x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment,
        ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
    }
    //Create certificate using templet
    derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
    if err != nil {
        return err
    }
    //pem encoding of certificate
    CAcertPem := string(pem.EncodeToMemory(
        &pem.Block{
            Type:  "CERTIFICATE",
            Bytes: derBytes,
        },
    ))
    fmt.Println(CAcertPem)

	file, err := os.Create(path+"ca_key.pem")
	if err != nil {
		return err
	}
	defer file.Close()
	file.Write(CAkeyPem)
	file, err = os.Create(path+"ca_cert.pem")
	if err != nil {
		return err
	}
	defer file.Close()
	file.Write([]byte(CAcertPem))

	return nil
}

func GenerateServerCertKey(domain string, CApath string, path string) error {
	//Load CA certificate
	CAcertPEM, err := os.ReadFile(CApath + "ca_cert.pem")
	if err != nil {
		return err
	}
	
	//Load CA private key
	CAkeyPEM, err := os.ReadFile(CApath + "ca_key.pem")
	if err != nil {
		return err
	}

	//Decode PEM to DER
	CAcertDER, _ := pem.Decode(CAcertPEM)
	CAkeyDER, _ := pem.Decode(CAkeyPEM)

	//Parse DER to certificate
	CAcert, err := x509.ParseCertificate(CAcertDER.Bytes)
	if err != nil {
		return err
	}

	//Parse DER to private key
	CAkey, err := x509.ParsePKCS1PrivateKey(CAkeyDER.Bytes)
	if err != nil {
		return err
	}

	//Generate private key
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}
	keyBytes := x509.MarshalPKCS1PrivateKey(key)
	// PEM encoding of private key
	keyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: keyBytes,
		},
	)
	fmt.Println(string(keyPem))

	notBefore := time.Now()
	notAfter := notBefore.Add(365*24*10*time.Hour)
	
	//Create certificate templet
	template := x509.Certificate{
		SerialNumber:          big.NewInt(0),
		Subject:               pkix.Name{CommonName: domain},
		SignatureAlgorithm:    x509.SHA256WithRSA,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyAgreement | x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}
	//Create certificate using templet
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, CAcert, &key.PublicKey, CAkey)
	if err != nil {
		return err
	}

	//pem encoding of certificate
	certPem := string(pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: derBytes,
		},
	))
	fmt.Println(certPem)
		
	file, err := os.Create(path+"server_key.pem")
	if err != nil {
		return err
	}
	defer file.Close()

	file.Write(keyPem)
	file, err = os.Create(path+"server_cert.pem")
	if err != nil {
		return err
	}
	defer file.Close()

	file.Write([]byte(certPem))

	return nil
}