use anyhow::{anyhow, Result};
use chacha20poly1305::{
    aead::{stream, Aead, NewAead},
    XChaCha20Poly1305,
};
use rand::{rngs::OsRng, RngCore};
use std::{
    fs::{self, File},
    io::{Read, Write},
    str,
};

pub enum Actions {
    Lock,
    Unlock,
}

pub fn new(path: &String, action: Actions) {
    match action {
        crate::lock::Actions::Lock => match lock(path.as_str()) {
            Ok(_) => println!("File encrypted"),
            Err(e) => eprintln!("File encryption error: {}", e),
        },
        crate::lock::Actions::Unlock => match unlock(path.as_str()) {
            Ok(_) => println!("File decrypted"),
            Err(e) => eprintln!("File decryption error: {}", e),
        },
    }
}

fn lock(path: &str) -> Result<(), anyhow::Error> {
    let kryptfile = filename_as_kryptfile(path);
    let metadata = fs::metadata(path)?;
    if metadata.len() < 500 {
        encrypt_small_files(path, kryptfile.as_str())
    } else {
        encrypt_large_files(path, kryptfile.as_str())
    }
}

fn unlock(path: &str) -> Result<(), anyhow::Error> {
    let metadata = fs::metadata(path)?;

    let mut new_name: Vec<&str> = path.split(".").collect();
    for i in 0..new_name.len() {
        if new_name[i] == "krypt" && i > 0 {
            new_name[i] = "";
        }
    }
    let mut out_nc: Vec<String> = vec![];
    for l in new_name {
        let s = String::from(l);
        out_nc.push(s);
    }
    let mut out: String = String::from("");
    let mut instance = 0;
    for n in out_nc {
        if instance >= 1 && n != "" {
            out += format!(".{}", n).as_str();
        } else {
            out += &n;
        }
        instance += 1
    }

    if metadata.len() < 500 {
        decrypt_small_files(path, &out)
    } else {
        decrypt_large_files(path, &out)
    }
}

fn encrypt_small_files(filepath: &str, dist: &str) -> Result<(), anyhow::Error> {
    let mut key = [0u8; 32];
    let mut nonce = [0u8; 24];
    OsRng.fill_bytes(&mut key);
    OsRng.fill_bytes(&mut nonce);
    let cipher = XChaCha20Poly1305::new(&key.into());
    let file_data = fs::read(filepath)?;

    let encrypted_file = cipher
        .encrypt(&nonce.into(), file_data.as_ref())
        .map_err(|err| anyhow!("Encrypting small files: {}", err))?;

    fs::write(&dist, encrypted_file)?;
    let mut kn: Vec<u8> = vec![];
    kn.extend_from_slice(&key);
    kn.extend_from_slice(&nonce);
    fs::write("kryptfle", kn)?;

    Ok(())
}

fn filename_as_kryptfile(path: &str) -> String {
    let result = format!("{}.krypt", path);
    result
}

fn decrypt_small_files(encrypted_file_path: &str, dist: &str) -> Result<(), anyhow::Error> {
    if let Ok(k) = fetch_key("kryptfle") {
        let cipher = XChaCha20Poly1305::new(&k.0.into());

        let file_data = fs::read(encrypted_file_path)?;

        let decrypted_file = cipher
            .decrypt(&k.1.into(), file_data.as_ref())
            .map_err(|err| anyhow!("Decrypting small file: {}", err))?;

        fs::write(&dist, decrypted_file)?;
    }

    Ok(())
}

fn encrypt_large_files(filepath: &str, dist: &str) -> Result<(), anyhow::Error> {
    let mut key = [0u8; 32];
    let mut nonce = [0u8; 19];
    OsRng.fill_bytes(&mut key);
    OsRng.fill_bytes(&mut nonce);
    let aead = XChaCha20Poly1305::new(&key.into());
    let mut stream_encryptor = stream::EncryptorBE32::from_aead(aead, &nonce.into());

    const BUFFER_LEN: usize = 500;
    let mut buffer = [0u8; BUFFER_LEN];

    let mut source_file = File::open(filepath)?;
    let mut dist_file = File::create(dist)?;

    let mut kn: Vec<u8> = vec![];
    kn.extend_from_slice(&key);
    kn.extend_from_slice(&nonce);
    fs::write("kryptfle", kn)?;

    loop {
        let read_count = source_file.read(&mut buffer)?;
        if read_count == BUFFER_LEN {
            let ciphertext = stream_encryptor
                .encrypt_next(buffer.as_slice())
                .map_err(|err| anyhow!("Encrypting large file: {}", err))?;
            dist_file.write(&ciphertext)?;
        } else {
            let ciphertext = stream_encryptor
                .encrypt_last(&buffer[..read_count])
                .map_err(|err| anyhow!("Encrypting large file: {}", err))?;
            dist_file.write(&ciphertext)?;
            break;
        }
    }

    Ok(())
}

fn decrypt_large_files(encrypted_file_path: &str, dist: &str) -> Result<(), anyhow::Error> {
    if let Ok(k) = fetch_large_key("kryptfle") {
        let aead = XChaCha20Poly1305::new(k.0.as_ref().into());
        let mut stream_decryptor = stream::DecryptorBE32::from_aead(aead, k.1.as_ref().into());

        const BUFFER_LEN: usize = 500 + 16;
        let mut buffer = [0u8; BUFFER_LEN];

        let mut encrypted_file = File::open(encrypted_file_path)?;
        let mut dist_file = File::create(dist)?;
        loop {
            let read_count = encrypted_file.read(&mut buffer)?;

            if read_count == BUFFER_LEN {
                let plaintext = stream_decryptor
                    .decrypt_next(buffer.as_slice())
                    .map_err(|err| anyhow!("Decrypting large file: {}", err))?;
                dist_file.write(&plaintext)?;
            } else if read_count == 0 {
                break;
            } else {
                let plaintext = stream_decryptor
                    .decrypt_last(&buffer[..read_count])
                    .map_err(|err| anyhow!("Decrypting large file: {}", err))?;
                dist_file.write(&plaintext)?;
                break;
            }
        }
    }

    Ok(())
}

fn fetch_key(path: &str) -> Result<([u8; 32], [u8; 24]), anyhow::Error> {
    if let Ok(s) = fs::read(path) {
        let split: Vec<&[u8]> = s.chunks(32).collect();

        let old_key = |slice: &[u8]| -> [u8; 32] {
            slice[..]
                .try_into()
                .expect("failed to grab key from kryptfle")
        };
        let old_nonce = |slice: &[u8]| -> [u8; 24] {
            slice[..]
                .try_into()
                .expect("failed to grab key from kryptfle")
        };

        let k = old_key(split[0]);
        let ntwice = old_nonce(split[1]);
        Ok((k, ntwice))
    } else {
        panic!("Unable to read kryptfle.")
    }
}

fn fetch_large_key(path: &str) -> Result<([u8; 32], [u8; 19]), anyhow::Error> {
    if let Ok(s) = fs::read(path) {
        let split: Vec<&[u8]> = s.chunks(32).collect();

        let old_key = |slice: &[u8]| -> [u8; 32] {
            slice[..]
                .try_into()
                .expect("failed to grab key from kryptfle")
        };
        let old_nonce = |slice: &[u8]| -> [u8; 19] {
            slice[..]
                .try_into()
                .expect("failed to grab key from kryptfle")
        };

        let k = old_key(split[0]);
        let ntwice = old_nonce(split[1]);
        Ok((k, ntwice))
    } else {
        panic!("Unable to read kryptfle.")
    }
}
