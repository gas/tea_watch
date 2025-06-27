![logo_tea_watch](https://github.com/user-attachments/assets/ec88ee38-1b54-40a9-9a38-fa18c29b97a1)

![Go Build & Test](https://github.com/gas/tea_watch/actions/workflows/go.yml/badge.svg) ![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg) ![GitHub release (latest by date)](https://img.shields.io/github/v/release/gas/tea_watch) | [README en español](README.md)

`tea_watch` is a terminal utility, written in Go with Lipgloss, for monitoring filesystem changes in real-time. It's very useful for any process that starts modifying files (I'm not looking at anyone, Gemini-cli...).

![Screenshot of tea_watch in action](https://github.com/user-attachments/assets/fb7c343a-42cd-420c-bd1a-ff27900b8945?raw=true)

## Features

* **Real-Time Monitoring:** Uses `fsnotify` for efficient, native event detection.
* **Clear & Dynamic Interface:** Built with `Bubble Tea` and `Lipgloss`, the UI adapts to your terminal size.
* **Event Counters:** Visualize how many times a file has been created, written to, renamed, or deleted.
* **Intuitive Navigation:** Move through the file list with the keyboard arrows or the mouse wheel.
* **Real-Time Filtering:** Press `/` to start typing and instantly filter the file list.
* **Event Highlighting:** Files with recent changes are subtly highlighted to draw your attention.
* **Intelligent Management:** Groups "atomic" save events and hides deleted files after a timeout to keep the view clean.

## Requirements

* **(Recommended)** A [Nerd Font](https://www.nerdfonts.com/) installed and configured in your terminal to correctly display icons.
* If you don't use a Nerd Font, don't worry! The application will run in an alternative ASCII mode.

## Installation

You have several methods to install `tea_watch`. Choose the one that best suits you.

### Method 1: Automatic Script (Linux & macOS)

This is the easiest and fastest way. Simply copy and paste this line into your terminal. The script will detect your OS, download the latest version, install it, and attempt to set up a keyboard shortcut (`Alt+W`) automatically.

```bash
curl -sSL https://raw.githubusercontent.com/gas/tea_watch/main/install.sh | bash
```
After it finishes, restart your terminal or run `source ~/.bashrc` (or `~/.zshrc`).

### Method 2: With `go install` (if you have Go installed)

This method compiles and installs the binary, but **it does not configure the keyboard shortcut**.

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
# Monitor the current directory
tea_watch

# Force ASCII mode (no icons)
tea_watch --no-nerd-fonts

# Monitor a specific directory
tea_watch /path/to/your/directory

# Use the flag and a directory at the same time
tea_watch --no-nerd-fonts /path/to/your/directory
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