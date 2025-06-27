![logo_tea_watch](https://github.com/user-attachments/assets/ec88ee38-1b54-40a9-9a38-fa18c29b97a1)

![Go Build & Test](https://github.com/gas/tea_watch/actions/workflows/go.yml/badge.svg) ![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg) ![GitHub release (latest by date)](https://img.shields.io/github/v/release/gas/tea_watch)
[README en español](README.md)

`tea_watch` is a terminal utility, written in Go with Lipgloss, for monitoring filesystem changes in real-time. Check and count file modifications at current directory.

![teawatch in action](https://github.com/user-attachments/assets/cc4520f1-454f-4124-8c7d-477d4697807f?raw=true)


## Features

* **Real-Time Monitoring:** Uses `fsnotify` for efficient, native event detection.
* **Slick Interface:** TUI built with `Bubble Tea` and `Lipgloss`, fits your terminal size.
* **Event Counters:** Visualize and count file modifications since created (written to or renamed).
* **Navigation:** Move through the file list with the keyboard arrows or mouse wheel.
* **Filtering:** Press `/` to filter the file list by query.
* **Event Highlighting:** For easy noting the new ones.
* **Discard atomic events:** Ephemeral file saving events are counted but hidden for clarity. Also hides deleted files after a timeout.


## Requirements

* **(Recommended)** A [Nerd Font](https://www.nerdfonts.com/) installed and configured in your terminal to correctly display icons.
* If you don't use a Nerd Font, it can run in ASCII mode (flag *--nerd-fonts=false*).


## Installation

You have several methods to install `tea_watch`.

### Method 1: Automatic Script (Linux & macOS)

Easiest and fastest. Simply copy and paste this line into your terminal. The script will detect your OS, download the latest version, install it, write a default config.toml file (at ~/.config/tea_watch) and attempt to set up a keyboard shortcut (`Alt+W`) automatically.

```bash
curl -sSL https://raw.githubusercontent.com/gas/tea_watch/main/install.sh | bash
```
After it finishes, restart your terminal or run `source ~/.bashrc` (or `~/.zshrc`).

### Method 2: With `go install` (if you have Go installed)

This method compiles and installs the binary, but **it does not configure the keyboard shortcut** or create the configuration file. You can copy the [config.example.toml](config.example.toml) from this repository.

```bash
go install github.com/gas/tea_watch@latest
```
After installation, you can run the tool with the `tea_watch` command. If you want the keyboard shortcut, follow the steps in "Manual Shortcut Configuration" below.

### Method 3: Pre-compiled Binaries (Manual)

This method gives you full control. You download the binary and place it wherever you want. **It does not configure the keyboard shortcut**.

1.  Download the file for your system from the [Releases page](https://github.com/gas/tea_watch/releases).
2.  Unzip it.
3.  Make the file executable (on Linux & macOS): `chmod +x tea_watch`
4.  (Recommended) Move the file to a directory in your `$PATH`, for example: `sudo mv tea_watch /usr/local/bin/`

---


## Usage

Simply run the command in your terminal:

```bash
# Monitor the current directory (will use your config.toml)
tea_watch

# Force ASCII mode (no NF icons, will ignore config.toml)
tea_watch --nerd-fonts=false

# Monitor a specific directory
tea_watch /path/to/your/directory
```

### Keybindings

| Key(s)         | Action                               |
| -------------- | ------------------------------------ |
| `↑` / `k`      | Move cursor up                       |
| `↓` / `j`      | Move cursor down                     |
| `Mouse Wheel`  | Scroll through the list              |
| `/`            | Enter/exit filter mode               |
| `Esc`          | Exit filter mode / Exit the program  |
| `q` / `Ctrl+C` | Exit the program                     |


## Localization (Translation)

You can translate `tea_watch` to any language.

1. After installing the application, look for the configuration file that has been created at:
 `~/.config/tea_watch/config.toml`
 It should be the same as [config.example.toml](config.example.toml) in this repository.

2.  Open the file with a text editor. You will see a `[strings]` section with all the English phrases commented out.

3.  Uncomment the lines and translate them if you want. French example:

    ```toml
 [strings]
 monitoring = "Surveillance"
 filter_prompt = "Filtrer: "
 total_events = "Événements"
    # ...etc.
 ````
