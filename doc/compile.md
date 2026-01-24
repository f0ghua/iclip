# iclip GUI 编译指南

## 环境要求

### Linux (Ubuntu) 开发环境

1. **安装 Go 语言环境**
   ```bash
   # 下载 Go (版本 1.21+)
   wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
   sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

   # 添加到 PATH
   echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
   source ~/.bashrc

   # 验证安装
   go version
   ```

2. **安装 Fyne 依赖**
   ```bash
   # Fyne 需要的系统库
   sudo apt-get update
   sudo apt-get install -y gcc libgl1-mesa-dev xorg-dev
   ```

3. **安装运行时依赖（仅 Linux 客户端需要）**
   ```bash
   # 剪贴板访问
   sudo apt-get install -y xclip

   # SMB 文件传输
   sudo apt-get install -y smbclient
   ```

### Windows 开发环境（可选）

如果需要在 Windows 上本地编译：

1. **安装 Go**
   - 下载: https://go.dev/dl/
   - 运行安装程序

2. **安装 GCC (MinGW-w64)**
   ```powershell
   # 使用 Chocolatey
   choco install mingw

   # 或下载 MSYS2: https://www.msys2.org/
   ```

## 编译 iclip_client_gui (Linux 版本)

```bash
cd iclip_client_gui

# 下载依赖
go mod tidy

# 编译 Linux 版本
go build -o iclip_client_gui

# 或者编译带优化的版本（更小的二进制）
go build -ldflags="-s -w" -o iclip_client_gui
```

## 交叉编译 iclip_server_gui (Windows 版本)

在 Linux 上编译 Windows 可执行文件需要安装 MinGW-w64：

```bash
# 安装 MinGW-w64 交叉编译器
sudo apt-get install -y gcc-mingw-w64-x86-64

cd iclip_server_gui

# 下载依赖
go mod tidy

# 交叉编译 Windows 版本
CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 \
    go build -ldflags="-s -w -H=windowsgui" -o iclip_server_gui.exe
```

注意：
- `-H=windowsgui` 隐藏控制台窗口（GUI 应用推荐）
- `-ldflags="-s -w"` 去除调试信息，减小文件体积

## 在 Windows 上本地编译 iclip_server_gui

```powershell
cd iclip_server_gui

# 下载依赖
go mod tidy

# 编译
go build -ldflags="-s -w -H=windowsgui" -o iclip_server_gui.exe
```

## 部署

### iclip_client_gui (Linux)

1. 将 `iclip_client_gui` 可执行文件复制到目标目录
2. 在同目录创建 `config.ini` 配置文件
3. 确保安装了 `xclip` 和 `smbclient`
4. 运行: `./iclip_client_gui`

### iclip_server_gui (Windows)

1. 将 `iclip_server_gui.exe` 复制到目标目录
2. 在同目录创建 `config.ini` 配置文件
3. 双击运行或通过命令行启动

## 配置文件

### 客户端配置 (iclip_client_gui/config.ini)

```ini
[smb]
server = 172.21.67.235
share = Send2
path = PD-6-4
username = 11867
password = YOUR_PASSWORD
workgroup = sdc

[monitor]
interval = 1.0
```

### 服务端配置 (iclip_server_gui/config.ini)

```ini
[monitor]
watch_path = \\172.21.68.235\Receive2\PD-6-4
interval = 1.0
pattern = clipboard_*.tar.gz
```

## 故障排除

### Linux 客户端

1. **xclip 未找到**
   ```bash
   sudo apt-get install xclip
   ```

2. **smbclient 连接失败**
   - 检查网络连接
   - 检查 SMB 共享是否可访问
   - 验证用户名密码是否正确

3. **GUI 无法启动**
   - 确保有显示环境 (X11/Wayland)
   - 安装: `sudo apt-get install libgl1-mesa-dev`

### Windows 服务端

1. **无法访问网络路径**
   - 确保 UNC 路径格式正确: `\\server\share\path`
   - 检查网络权限

2. **剪贴板操作失败**
   - 检查是否有其他程序锁定剪贴板
   - 尝试重启应用

## 打包发布

### 创建发布包

```bash
# 创建发布目录
mkdir -p release/linux release/windows

# 复制 Linux 版本
cp iclip_client_gui/iclip_client_gui release/linux/
cp iclip_client_gui/config.ini.example release/linux/config.ini

# 复制 Windows 版本
cp iclip_server_gui/iclip_server_gui.exe release/windows/
cp iclip_server_gui/config.ini.example release/windows/config.ini

# 创建压缩包
cd release
tar -czf iclip_client_gui_linux_amd64.tar.gz linux/
zip -r iclip_server_gui_windows_amd64.zip windows/
```
