package tls

import (
	"fmt"
	"strings"
	"testing"

	r "github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestPrivateKeyRSA(t *testing.T) {
	r.Test(t, r.TestCase{
		Providers: testProviders,
		Steps: []r.TestStep{
			r.TestStep{
				Config: `
                    resource "tls_private_key" "test" {
                        algorithm = "RSA"
                    }
                    output "private_key_pem" {
                        value = "${tls_private_key.test.private_key_pem}"
                    }
                    output "public_key_pem" {
                        value = "${tls_private_key.test.public_key_pem}"
                    }
                    output "public_key_openssh" {
                        value = "${tls_private_key.test.public_key_openssh}"
                    }
                    output "public_key_fingerprint_md5" {
                        value = "${tls_private_key.test.public_key_fingerprint_md5}"
                    }
                `,
				Check: func(s *terraform.State) error {
					gotPrivateUntyped := s.RootModule().Outputs["private_key_pem"].Value
					gotPrivate, ok := gotPrivateUntyped.(string)
					if !ok {
						return fmt.Errorf("output for \"private_key_pem\" is not a string")
					}

					if !strings.HasPrefix(gotPrivate, "-----BEGIN RSA PRIVATE KEY----") {
						return fmt.Errorf("private key is missing RSA key PEM preamble")
					}

					if len(gotPrivate) > 1700 {
						return fmt.Errorf("private key PEM looks too long for a 2048-bit key (got %v characters)", len(gotPrivate))
					}

					gotPublicUntyped := s.RootModule().Outputs["public_key_pem"].Value
					gotPublic, ok := gotPublicUntyped.(string)
					if !ok {
						return fmt.Errorf("output for \"public_key_pem\" is not a string")
					}
					if !strings.HasPrefix(gotPublic, "-----BEGIN PUBLIC KEY----") {
						return fmt.Errorf("public key is missing public key PEM preamble")
					}

					gotPublicSSHUntyped := s.RootModule().Outputs["public_key_openssh"].Value
					gotPublicSSH, ok := gotPublicSSHUntyped.(string)
					if !ok {
						return fmt.Errorf("output for \"public_key_openssh\" is not a string")
					}
					if !strings.HasPrefix(gotPublicSSH, "ssh-rsa ") {
						return fmt.Errorf("SSH public key is missing ssh-rsa prefix")
					}

					gotPublicFingerprintUntyped := s.RootModule().Outputs["public_key_fingerprint_md5"].Value
					gotPublicFingerprint, ok := gotPublicFingerprintUntyped.(string)
					if !ok {
						return fmt.Errorf("output for \"public_key_fingerprint_md5\" is not a string")
					}
					if !(gotPublicFingerprint[2] == ':') {
						return fmt.Errorf("MD5 public key fingerprint is missing : in the correct place")
					}

					return nil
				},
			},
			r.TestStep{
				Config: `
                    resource "tls_private_key" "test" {
                        algorithm = "RSA"
                        rsa_bits = 4096
                    }
                    output "key_pem" {
                        value = "${tls_private_key.test.private_key_pem}"
                    }
                `,
				Check: func(s *terraform.State) error {
					gotUntyped := s.RootModule().Outputs["key_pem"].Value
					got, ok := gotUntyped.(string)
					if !ok {
						return fmt.Errorf("output for \"key_pem\" is not a string")
					}
					if !strings.HasPrefix(got, "-----BEGIN RSA PRIVATE KEY----") {
						return fmt.Errorf("key is missing RSA key PEM preamble")
					}
					if len(got) < 1700 {
						return fmt.Errorf("key PEM looks too short for a 4096-bit key (got %v characters)", len(got))
					}
					return nil
				},
			},
		},
	})
}

func TestPrivateKeyECDSA(t *testing.T) {
	r.Test(t, r.TestCase{
		Providers: testProviders,
		Steps: []r.TestStep{
			r.TestStep{
				Config: `
                    resource "tls_private_key" "test" {
                        algorithm = "ECDSA"
                    }
                    output "private_key_pem" {
                        value = "${tls_private_key.test.private_key_pem}"
                    }
                    output "public_key_pem" {
                        value = "${tls_private_key.test.public_key_pem}"
                    }
                    output "public_key_openssh" {
                        value = "${tls_private_key.test.public_key_openssh}"
                    }
                    output "public_key_fingerprint_md5" {
                        value = "${tls_private_key.test.public_key_fingerprint_md5}"
                    }
                `,
				Check: func(s *terraform.State) error {
					gotPrivateUntyped := s.RootModule().Outputs["private_key_pem"].Value
					gotPrivate, ok := gotPrivateUntyped.(string)
					if !ok {
						return fmt.Errorf("output for \"private_key_pem\" is not a string")
					}

					if !strings.HasPrefix(gotPrivate, "-----BEGIN EC PRIVATE KEY----") {
						return fmt.Errorf("Private key is missing EC key PEM preamble")
					}

					gotPublicUntyped := s.RootModule().Outputs["public_key_pem"].Value
					gotPublic, ok := gotPublicUntyped.(string)
					if !ok {
						return fmt.Errorf("output for \"public_key_pem\" is not a string")
					}

					if !strings.HasPrefix(gotPublic, "-----BEGIN PUBLIC KEY----") {
						return fmt.Errorf("public key is missing public key PEM preamble")
					}

					gotPublicSSH := s.RootModule().Outputs["public_key_openssh"].Value.(string)
					if gotPublicSSH != "" {
						return fmt.Errorf("P224 EC key should not generate OpenSSH public key")
					}

					gotPublicFingerprint := s.RootModule().Outputs["public_key_fingerprint_md5"].Value.(string)
					if gotPublicFingerprint != "" {
						return fmt.Errorf("P224 EC key should not generate OpenSSH public key fingerprint")
					}

					return nil
				},
			},
			r.TestStep{
				Config: `
                    resource "tls_private_key" "test" {
                        algorithm = "ECDSA"
                        ecdsa_curve = "P256"
                    }
                    output "private_key_pem" {
                        value = "${tls_private_key.test.private_key_pem}"
                    }
                    output "public_key_pem" {
                        value = "${tls_private_key.test.public_key_pem}"
                    }
                    output "public_key_openssh" {
                        value = "${tls_private_key.test.public_key_openssh}"
                    }
                    output "public_key_fingerprint_md5" {
                        value = "${tls_private_key.test.public_key_fingerprint_md5}"
                    }
                `,
				Check: func(s *terraform.State) error {
					gotPrivateUntyped := s.RootModule().Outputs["private_key_pem"].Value
					gotPrivate, ok := gotPrivateUntyped.(string)
					if !ok {
						return fmt.Errorf("output for \"private_key_pem\" is not a string")
					}
					if !strings.HasPrefix(gotPrivate, "-----BEGIN EC PRIVATE KEY----") {
						return fmt.Errorf("Private key is missing EC key PEM preamble")
					}

					gotPublicUntyped := s.RootModule().Outputs["public_key_pem"].Value
					gotPublic, ok := gotPublicUntyped.(string)
					if !ok {
						return fmt.Errorf("output for \"public_key_pem\" is not a string")
					}
					if !strings.HasPrefix(gotPublic, "-----BEGIN PUBLIC KEY----") {
						return fmt.Errorf("public key is missing public key PEM preamble")
					}

					gotPublicSSHUntyped := s.RootModule().Outputs["public_key_openssh"].Value
					gotPublicSSH, ok := gotPublicSSHUntyped.(string)
					if !ok {
						return fmt.Errorf("output for \"public_key_openssh\" is not a string")
					}
					if !strings.HasPrefix(gotPublicSSH, "ecdsa-sha2-nistp256 ") {
						return fmt.Errorf("P256 SSH public key is missing ecdsa prefix")
					}

					gotPublicFingerprintUntyped := s.RootModule().Outputs["public_key_fingerprint_md5"].Value
					gotPublicFingerprint, ok := gotPublicFingerprintUntyped.(string)
					if !ok {
						return fmt.Errorf("output for \"public_key_fingerprint_md5\" is not a string")
					}
					if !(gotPublicFingerprint[2] == ':') {
						return fmt.Errorf("MD5 public key fingerprint is missing : in the correct planbe")
					}

					return nil
				},
			},
		},
	})
}

func TestPrivateKeyEd25519(t *testing.T) {
	r.Test(t, r.TestCase{
		Providers: testProviders,
		Steps: []r.TestStep{
			r.TestStep{
				Config: `
                    resource "tls_private_key" "test" {
                    	algorithm = "ED25519"
                    }
                    output "private_key_pem" {
                    	value = "${tls_private_key.test.private_key_pem}"
                    }
                    output "public_key_pem" {
                    	value = "${tls_private_key.test.public_key_pem}"
                    }
                    output "public_key_openssh" {
                    	value = "${tls_private_key.test.public_key_openssh}"
                    }
                    output "public_key_fingerprint_md5" {
                    	value = "${tls_private_key.test.public_key_fingerprint_md5}"
                    }
                    `,
				Check: func(s *terraform.State) error {
					gotPrivateUntyped := s.RootModule().Outputs["private_key_pem"].Value
					gotPrivate, ok := gotPrivateUntyped.(string)
					if !ok {
						return fmt.Errorf("output for \"private_key_pem\" is not a string")
					}
					if !strings.HasPrefix(gotPrivate, "-----BEGIN ED25519 PRIVATE KEY-----") {
						return fmt.Errorf("prviate key is missing ED25519 key PEM preamble")
					}
					if len(gotPrivate) > 115 {
						return fmt.Errorf("private key PEM looks too long for a 32-bit key (got %v characters)", len(gotPrivate))
					}

					gotPublicUntyped := s.RootModule().Outputs["public_key_pem"].Value
					gotPublic, ok := gotPublicUntyped.(string)
					if !ok {
						return fmt.Errorf("output for \"public_key_pem\" is not a string")
					}
					if !strings.HasPrefix(gotPublic, "-----BEGIN PUBLIC KEY----") {
						return fmt.Errorf("public key is missing public key PEM preamble")
					}

					gotPublicSSHUntyped := s.RootModule().Outputs["public_key_openssh"].Value
					gotPublicSSH, ok := gotPublicSSHUntyped.(string)
					if !ok {
						return fmt.Errorf("output for \"public_key_openssh\" is not a string")
					}
					if !strings.HasPrefix(gotPublicSSH, "ssh-ed25519") {
						return fmt.Errorf("SSH public key is missing ssh-ed25519 prefix")
					}

					gotPublicFingerprintUntyped := s.RootModule().Outputs["public_key_fingerprint_md5"].Value
					gotPublicFingerprint, ok := gotPublicFingerprintUntyped.(string)
					if !ok {
						return fmt.Errorf("output for \"public_key_fingerprint_md5\" is not a string")
					}
					if !(gotPublicFingerprint[2] == ':') {
						return fmt.Errorf("MD5 public key fingerprint is missing : in the correct place")
					}

					return nil
				},
			},
		},
	})
}
