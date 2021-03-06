package tls

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"golang.org/x/crypto/ed25519"
)

type keyAlgo func(d *schema.ResourceData) (interface{}, error)
type keyParser func([]byte) (interface{}, error)

var keyAlgos map[string]keyAlgo = map[string]keyAlgo{
	"RSA": func(d *schema.ResourceData) (interface{}, error) {
		rsaBits := d.Get("rsa_bits").(int)
		return rsa.GenerateKey(rand.Reader, rsaBits)
	},
	"ECDSA": func(d *schema.ResourceData) (interface{}, error) {
		curve := d.Get("ecdsa_curve").(string)
		switch curve {
		case "P224":
			return ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
		case "P256":
			return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		case "P384":
			return ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
		case "P521":
			return ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
		default:
			return nil, fmt.Errorf("invalid ecdsa_curve; must be P224, P256, P384 or P521")
		}
	},
	"ED25519": func(d *schema.ResourceData) (interface{}, error) {
		return func() (interface{}, error) {
			_, priv, err := ed25519.GenerateKey(rand.Reader)
			return priv, err
		}()
	},
}

var keyParsers map[string]keyParser = map[string]keyParser{
	"RSA": func(der []byte) (interface{}, error) {
		return x509.ParsePKCS1PrivateKey(der)
	},
	"ECDSA": func(der []byte) (interface{}, error) {
		return x509.ParseECPrivateKey(der)
	},
	"ED25519": func(der []byte) (interface{}, error) {
		return ed25519.NewKeyFromSeed(der), nil
	},
}

func resourcePrivateKey() *schema.Resource {
	return &schema.Resource{
		Create: CreatePrivateKey,
		Delete: DeletePrivateKey,
		Read:   ReadPrivateKey,

		Schema: map[string]*schema.Schema{
			"algorithm": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the algorithm to use to generate the private key",
				ForceNew:    true,
			},

			"rsa_bits": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Number of bits to use when generating an RSA key",
				ForceNew:    true,
				Default:     2048,
			},

			"ecdsa_curve": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ECDSA curve to use when generating a key",
				ForceNew:    true,
				Default:     "P224",
			},

			"private_key_pem": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"public_key_pem": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"public_key_openssh": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"public_key_fingerprint_md5": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func CreatePrivateKey(d *schema.ResourceData, meta interface{}) error {
	keyAlgoName := d.Get("algorithm").(string)
	var keyFunc keyAlgo
	var ok bool
	if keyFunc, ok = keyAlgos[keyAlgoName]; !ok {
		return fmt.Errorf("invalid key_algorithm %#v", keyAlgoName)
	}

	key, err := keyFunc(d)
	if err != nil {
		return err
	}

	var keyPemBlock *pem.Block
	switch k := key.(type) {
	case *rsa.PrivateKey:
		keyPemBlock = &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(k),
		}
	case *ecdsa.PrivateKey:
		keyBytes, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			return fmt.Errorf("error encoding key to PEM: %s", err)
		}
		keyPemBlock = &pem.Block{
			Type:  "EC PRIVATE KEY",
			Bytes: keyBytes,
		}
	case ed25519.PrivateKey:
		keyPemBlock = &pem.Block{
			Type:  "ED25519 PRIVATE KEY",
			Bytes: k.Seed(),
		}
	default:
		return fmt.Errorf("unsupported private key type")
	}
	keyPem := string(pem.EncodeToMemory(keyPemBlock))

	d.Set("private_key_pem", keyPem)
	return readPublicKey(d, key)
}

func DeletePrivateKey(d *schema.ResourceData, meta interface{}) error {
	d.SetId("")
	return nil
}

func ReadPrivateKey(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	case ed25519.PrivateKey:
		return k.Public()
	default:
		return nil
	}
}

func publicKeyBytes(priv interface{}) ([]byte, error) {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return x509.MarshalPKIXPublicKey(&k.PublicKey)
	case *ecdsa.PrivateKey:
		return x509.MarshalPKIXPublicKey(&k.PublicKey)
	case ed25519.PrivateKey:
		pubKey, ok := k.Public().(ed25519.PublicKey)
		if !ok {
			return nil, fmt.Errorf("failed to get ed25519 public key")
		}
		return []byte(pubKey), nil
	default:
		return nil, fmt.Errorf("unsupported private key type")
	}
}
