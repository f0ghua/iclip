# iclip - Cross-Environment Clipboard Transfer Tool

iclip is a tool for transferring clipboard content from an internal Ubuntu machine to an external Windows PC via SMB share, designed for restricted environments like Imera.

## System Architecture

```
┌─────────────────────┐         SMB          ┌─────────────────────┐
│   Ubuntu (Internal) │  ──────────────────► │   Windows (External)│
│                     │                      │                     │
│   iclip_client      │  clipboard_xxx.tar.gz│   iclip_server      │
│   - Monitor clipboard│ ──────────────────► │   - Monitor directory│
│   - Package & upload │  //172.21.67.235/   │   - Extract & set    │
│                     │   Send2/PD-6-4/      │     clipboard        │
└─────────────────────┘                      └─────────────────────┘
```

## Components

| Component | Platform | Language | Description |
|-----------|----------|----------|-------------|
| iclip_client | Ubuntu/Linux | C | Monitors clipboard, packages and uploads via SMB |
| iclip_server | Windows | Python 3 | Monitors shared directory, extracts and sets clipboard |

## Requirements

### Ubuntu (iclip_client)

- GCC compiler (for building)
- System tools:
  - `xclip` - clipboard operations
  - `smbclient` - SMB file transfer
  - `tar` - archive creation

```bash
sudo apt install xclip smbclient
```

### Windows (iclip_server)

- Python 3.8+
- pyperclip package

```cmd
pip install pyperclip
```

## Installation

### Ubuntu Client

1. Copy the `iclip_client` directory to the Ubuntu machine

2. Build the client:
   ```bash
   cd iclip_client
   make
   ```

3. Edit `config.ini` with your SMB settings:
   ```ini
   [smb]
   server = 172.21.67.235
   share = Send2
   path = PD-6-4
   username = 11867
   password = YOUR_PASSWORD_HERE
   workgroup = sdc

   [monitor]
   interval = 1.0
   ```

### Windows Server

1. Copy the `iclip_server` directory to the Windows machine

2. Install dependencies:
   ```cmd
   pip install pyperclip
   ```

3. Edit `config.ini` with your watch path:
   ```ini
   [monitor]
   watch_path = \\172.21.68.235\Receive2\PD-6-4
   interval = 1.0
   pattern = clipboard_*.tar.gz
   ```

## Usage

### Start the Server (Windows)

```cmd
cd iclip_server
python iclip_server.py
```

### Start the Client (Ubuntu)

```bash
cd iclip_client
./iclip_client
```

### Keyboard Controls

| Key | Action |
|-----|--------|
| `q` | Quit |
| `p` | Pause/Resume monitoring |
| `r` | Force refresh (client only) |

## How It Works

1. **Client (Ubuntu)**: Monitors the system clipboard every second (configurable)
2. When clipboard content changes:
   - Creates a `clipboard_data/` directory with:
     - `meta.json` - metadata (type, timestamp, username)
     - `content.txt` - the clipboard text
   - Packages into `clipboard_<username>.tar.gz`
   - Uploads to SMB share using `smbclient`
3. **Server (Windows)**: Monitors the shared directory
4. When a new `clipboard_*.tar.gz` file appears:
   - Extracts the archive
   - Reads the content
   - Sets it to Windows clipboard
   - Deletes the processed file

## Data Format

### Archive Structure

```
clipboard_<username>.tar.gz
└── clipboard_data/
    ├── meta.json
    └── content.txt
```

### meta.json Example

```json
{
  "type": "text",
  "timestamp": "2026-01-24T10:30:00",
  "username": "fog",
  "encoding": "utf-8"
}
```

## Troubleshooting

### SMB Connection Failed

1. Test SMB connection manually:
   ```bash
   smbclient '//172.21.67.235/Send2' -U '11867%password' -W sdc
   ```

2. Check if workgroup is required (add `-W <workgroup>` or set in config.ini)

3. Verify credentials and network connectivity

### Clipboard Not Detected (Ubuntu)

1. Ensure xclip is installed:
   ```bash
   which xclip
   ```

2. Test clipboard access:
   ```bash
   echo "test" | xclip -selection clipboard
   xclip -selection clipboard -o
   ```

### Server Not Receiving Files

1. Check if the watch path exists and is accessible
2. Verify the file pattern matches (`clipboard_*.tar.gz`)
3. Check Windows firewall settings

## Limitations

- Text content only (no images or files)
- One-way transfer (Ubuntu → Windows)
- Requires SMB share access

## License

MIT License
