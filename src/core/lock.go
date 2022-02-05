package core

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
	"log"
	"os"
)

type Action int

const (
	Lock Action = iota
	Unlock
)

type krypt struct {
	Input     string
	Output    string
	Action    Action
	Kryptfile string
}

func New() *krypt {
	return &krypt{
		Input:     "",
		Output:    "",
		Action:    0,
		Kryptfile: "kryptfile",
	}
}

func (k krypt) Start() error {
	target, err := os.Stat(k.Input)
	if os.IsNotExist(err) {
		return err
	}

	switch k.Action {
	case Lock:
		if target.Size() > 1e6 {
			return k.encrypt_large_file()
		}
		return k.lock()
	case Unlock:
		if target.Size() > 1e6 {
			return k.decrypt_large_file()
		}
		return k.unlock()
	default:
		if target.Size() > 1e6 {
			return k.encrypt_large_file()
		}
		return k.lock()
	}
}

func (k *krypt) lock() error {

	key := gen_key()

	file, err := os.ReadFile(k.Input)
	if err != nil {
		return err
	}

	ch := make(chan []byte)
	go HashFile(k.Input, ch)

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	if err := k.writeKey(key); err != nil {
		return err
	}

	ciphertext := gcm.Seal(nonce, nonce, file, nil)
	hash := <-ch
	secret := make([]byte, 0)
	secret = append(secret, hash...)
	secret = append(secret, ciphertext...)

	if err := os.WriteFile(k.Input+".krypt", secret, 0777); err != nil {
		return err
	}

	return nil
}

func (k *krypt) unlock() error {

	file, err := os.ReadFile(k.Input)
	if err != nil {
		return err
	}

	key, err := os.ReadFile("kryptfile")
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	hash := file[:32]
	rest := file[32:]

	nonce := rest[:gcm.NonceSize()]
	file = rest[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, file, nil)
	if err != nil {
		return err
	}

	k.gen_outfile_name()

	if err = os.WriteFile(k.Output, plaintext, 0777); err != nil {
		return err
	}
	ch := make(chan []byte)
	go HashFile(k.Output, ch)
	outhash := <-ch

	if !bytes.Equal(hash, outhash) {
		return errors.New("file does not share hash with kryptfile")
	}
	return nil

}

func (k *krypt) encrypt_large_file() error {

	infile, err := os.Open(k.Input)
	if err != nil {
		return err
	}
	defer infile.Close()

	key := gen_key()

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	iv := make([]byte, block.BlockSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		log.Fatal(err)
	}

	outfile, err := os.OpenFile(k.Input+".krypt", os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer outfile.Close()

	buf := make([]byte, 1024)
	stream := cipher.NewCTR(block, iv)
	for {
		n, err := infile.Read(buf)
		if n > 0 {
			stream.XORKeyStream(buf, buf[:n])
			outfile.Write(buf[:n])
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Printf("Read %d bytes: %v", n, err)
			break
		}
	}
	k.writeKey(key)
	outfile.Write(iv)

	return nil
}

func (k *krypt) decrypt_large_file() error {
	infile, err := os.Open(k.Input)
	if err != nil {
		return err
	}
	defer infile.Close()

	key, err := os.ReadFile("kryptfile")
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	fi, err := infile.Stat()
	if err != nil {
		return err
	}

	iv := make([]byte, block.BlockSize())
	msgLen := fi.Size() - int64(len(iv))
	_, err = infile.ReadAt(iv, msgLen)
	if err != nil {
		return err
	}
	k.gen_outfile_name()
	outfile, err := os.OpenFile(k.Output, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return err
	}

	buf := make([]byte, 1024)
	stream := cipher.NewCTR(block, iv)
	for {
		n, err := infile.Read(buf)
		if n > 0 {
			if n > int(msgLen) {
				n = int(msgLen)
			}
			msgLen -= int64(n)
			stream.XORKeyStream(buf, buf[:n])
			outfile.Write(buf[:n])

		}

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Printf("Read %d bytes: %v", n, err)
			break
		}

	}

	return nil
}
