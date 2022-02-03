use commands::{lock, lock::Actions};
use std::env;
mod commands;

fn main() {
    let args: Vec<String> = env::args().collect();
    
        match args[1].as_str() {
            "lock" | "l" => {
                if args.len() > 2 {
                    lock::new(&args[2], Actions::Lock);
                }
                
            }
            "unlock" | "u" => {
                if args.len() > 2 {
                    lock::new(&args[2], Actions::Unlock)
                }
                
            }
            _ => println!("No argument supplied"),
        }
    
}
