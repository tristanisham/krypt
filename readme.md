# Krypt
A simple encryption/compression tool written in Rust.

## Features
1. Encrypt files with ChaCha20 + nonce encryption
2. Decrypt files using kryptfile
    * Todo: Give users different methods for exporting kryptfiles

### Todo
* Add option to compress files using *snap* https://crates.io/crates/snap
* Add Multi-threading
* Add recursive encryption for directories

## About 
Krypt isn't unique. But it gets the job I needed it to do done. With it's very basic setup it's possible to infitly scale the program to use different encryption schemes, create private/public key pairs/ etc...
Generally, Krypt is for learning more about encryption (while still being perfectly servable for serious use).