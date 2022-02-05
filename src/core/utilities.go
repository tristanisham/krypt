package core

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)


func gen_key() []byte {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	return b
}

func (k *krypt) gen_outfile_name() {
	if strings.Contains(k.Input, ".krypt") {
		k.Output = strings.ReplaceAll(k.Input, ".krypt", "")
	} else {
		k.Output = k.Input
	}
}

func (k *krypt) writeKey(key []byte) error {
	if k.Kryptfile == "kryptfile" {
		if err := os.WriteFile(k.Kryptfile, key, 0666); err != nil {
			return err
		}
		return nil
	} else {
		if strings.Contains("http://", k.Kryptfile) || strings.Contains("https://", k.Kryptfile) {
			client := &http.Client{
				Timeout: time.Second * 10,
			}
			
			req, err := http.NewRequest("POST", k.Kryptfile, bytes.NewReader(key))
			if err != nil {
				return err
			}

			req.Header.Set("Content-Type", "application/octet-stream")
			rsp, _  := client.Do(req)
			if rsp.StatusCode != http.StatusOK {
				return fmt.Errorf("code: %s, failed to deliver kryptfile", rsp.Status)
			}
			
		}
	}
	return nil
}