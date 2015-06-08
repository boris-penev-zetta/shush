package main

import (
	"encoding/base64"
	"fmt"
	"git.realestate.com.au/mwilliams/shush/awsmeta"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/codegangsta/cli"
	"io/ioutil"
	"os"
)

func main() {

	app := cli.NewApp()
	app.Name = "shush"
	app.Version = "1.0.0"
	app.Usage = "KMS encryption and decryption"

	app.Commands = []cli.Command{
		{
			Name:  "encrypt",
			Usage: "Encrypt with a KMS key",
			Action: func(c *cli.Context) {
				if len(c.Args()) == 0 {
					abort(1, "no key specified")
				}
				key := c.Args().First()
				plaintext, err := getPayload(c.Args()[1:])
				if err != nil {
					abort(1, err)
				}
				ciphertext, err := encrypt(plaintext, key)
				if err != nil {
					abort(2, err)
				}
				fmt.Println(ciphertext)
			},
		},
		{
			Name:  "decrypt",
			Usage: "Decrypt KMS ciphertext",
			Action: func(c *cli.Context) {
				ciphertext, err := getPayload(c.Args())
				if err != nil {
					abort(1, err)
				}
				plaintext, err := decrypt(ciphertext)
				if err != nil {
					abort(2, err)
				}
				fmt.Print(plaintext)
			},
		},
	}

	app.Run(os.Args)

}

func kmsClient() *kms.KMS {
	region := os.Getenv("AWS_DEFAULT_REGION")
	if region == "" {
		region = awsmeta.GetRegion()
	}
	return kms.New(&aws.Config{Region: region})
}

func encrypt(plaintext string, key string) (string, error) {
	output, err := kmsClient().Encrypt(&kms.EncryptInput{
		KeyID:     &key,
		Plaintext: []byte(plaintext),
	})
	if err != nil {
		return "", err
	}
	ciphertext := base64.StdEncoding.EncodeToString(output.CiphertextBlob)
	return ciphertext, nil
}

func decrypt(ciphertext string) (string, error) {
	ciphertextBlob, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	output, err := kmsClient().Decrypt(&kms.DecryptInput{
		CiphertextBlob: ciphertextBlob,
	})
	if err != nil {
		return "", err
	}
	return string(output.Plaintext), nil
}

func getPayload(args []string) (string, error) {
	if len(args) >= 1 {
		return args[0], nil
	} else {
		input, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		return string(input), nil
	}
}

func abort(status int, message interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: %s", message)
	os.Exit(status)
}