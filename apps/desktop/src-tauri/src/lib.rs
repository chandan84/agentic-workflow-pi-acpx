//! AWPA desktop shell.
//!
//! The shell stays thin: the React app talks to the gateway-service over
//! Connect-Web directly. IPC commands exist only for things the web layer
//! cannot do — reading the local shell config and (later) secure token
//! storage and sidecar lifecycle for a bundled gateway.

use serde::Serialize;

#[derive(Serialize)]
pub struct ShellConfig {
    /// Gateway endpoint the web layer should use. Overridable with
    /// AWPA_GATEWAY_URL when launching the shell.
    pub gateway_url: String,
    pub shell_version: String,
}

#[tauri::command]
fn shell_config() -> ShellConfig {
    ShellConfig {
        gateway_url: std::env::var("AWPA_GATEWAY_URL")
            .unwrap_or_else(|_| "http://localhost:7000".to_string()),
        shell_version: env!("CARGO_PKG_VERSION").to_string(),
    }
}

/// Store an auth token in OS-native secure storage.
#[tauri::command]
fn store_token(_token: String) -> Result<(), String> {
    // TODO: persist via OS keychain (tauri-plugin-stronghold or keyring).
    Err("secure token storage not yet implemented".to_string())
}

pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .invoke_handler(tauri::generate_handler![shell_config, store_token])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
