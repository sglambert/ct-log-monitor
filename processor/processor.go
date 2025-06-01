package processor

import (
	"crypto/x509"
	"encoding/base64"
	"fmt"

	"github.com/google/certificate-transparency-go"
	"github.com/google/certificate-transparency-go/tls"
	"github.com/sglambert/ct-log-monitor/models"
)

const (
	X509EntryType    = 0
	PrecertEntryType = 1
)

func ParseLogEntry(entry models.LogEntry) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate

	// Decode leaf_input
	leafBytes, err := base64.StdEncoding.DecodeString(entry.LeafInput)
	if err != nil {
		fmt.Println("failed to base64 decode leaf_input: %w", err)
	}

	// Decode extra_data (it could be empty for X.509 entries)
	extraBytes, err := base64.StdEncoding.DecodeString(entry.ExtraData)
	if err != nil {
		fmt.Println("failed to base64 decode extra_data: %w", err)
	}

	// Parse raw leaf
	var leaf ct.MerkleTreeLeaf
	if rest, err := tls.Unmarshal(leafBytes, &leaf); err != nil {
		fmt.Println("failed to unmarshal MerkleTreeLeaf: %w", err)
	} else if len(rest) > 0 {
		fmt.Println("unexpected trailing data after MerkleTreeLeaf")
	}

	switch leaf.TimestampedEntry.EntryType {
	case ct.X509LogEntryType:
		// Single leaf certificate
		leafCert, err := x509.ParseCertificate(leaf.TimestampedEntry.X509Entry.Data)
		if err != nil {
			fmt.Println("failed to parse X.509 certificate: %w", err)
		}
		certs = append(certs, leafCert)

		var certChain []ct.ASN1Cert
		if rest, err := tls.Unmarshal(extraBytes, &certChain); err == nil && len(rest) == 0 {

			for _, asn1Cert := range certChain {
				cert, err := x509.ParseCertificate(asn1Cert.Data)
				if err == nil {
					certs = append(certs, cert)
				}
			}
		}

	default:
		return nil, nil
	}

	return certs, nil
}
