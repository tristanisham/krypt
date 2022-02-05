# Krypt
*v0.2.0*
A simple encryption/compression tool written in Go (formerly Rust).

## Features
1. Encrypt files with AES256 + nonce encryption
2. Verify authenticity of decrypted files with Sha256 hashing
2. Decrypt files using kryptfile
    * Todo: Give users different methods for exporting kryptfiles

### Todo
* Add option to compress files
* <s>Add Multi-threading</s>
* Add recursive encryption for directories

## About 
Krypt isn't unique. But it gets the job I needed it to do done. With it's very basic setup it's possible to infitly scale the program to use different encryption schemes, create private/public key pairs/ etc...
Generally, Krypt is for learning more about encryption (while still being perfectly servable for serious use).

## How to use
### Encrypt File
`$ krypt l -i bigfile.txt`

Use `l` or `lock` followed by a file path to encrypt the file. 
* Krypt will generate an encrypted `.krypt` file that shares the name of the encrypted file.
* Krypt will also generate a `kryptfile`. This file contains the **private key** for your keyfile, so don't lose it.

### Decrypt File
`$ krypt u -i bigfile.txt.krypt`

Use `u` or `unlock` in the same directory as your keyfile and krypt will decrypt and create the regular file without the `.krypt` extension. 