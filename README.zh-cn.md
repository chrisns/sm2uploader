[English Readme](README.md)

# SM2Uploader
Luban 和 Cura with SnapmakerPlugin 对于新手很友好，但是我的大部分配置文件在 PrusaSlicer 中，切片后再使用 Luban 上传到打印机是非常低效的。
这个工具提供了一步上传的能力，你可以通过命令行一次上传多个 gcode/cnc/bin固件 等文件。

## 功能
- 自动发现局域网内所有的 Snapmaker 打印机（和 Luban 相同的协议，使用 UDP 广播）
- 模拟 OctoPrint Server，这样就可以在各种切片软件，比如 Cura/PrusaSlicer/SuperSlicer/OrcaSlicer 中向 Snapmaker 打印机发送文件
- 为多挤出机提供智能预热、关闭不再使用的喷头等优化功能
- 强化擦料塔，避免多材料打印时因不粘合而倒塌，例如在 PETG+PLA 混合打印时
- Snapmaker 2 A-Series 第一次连接时需要授权，之后可以直接一步上传
- 支持 Snapmaker 2 A/J1/Artisan 全系列打印机
- 支持 macOS/Windows/Linux/RaspberryPi 多个平台

## 使用方法
下载适用的[程序文件](https://github.com/macdylan/sm2uploader/releases)
  - Linux/macOS 下，可能需要赋予可执行权限 `chmod +x sm2uploader`

Windows 使用方法:
  - 在解压目录中双击 `start-octoprint.bat` 批处理程序
  - 按照提示输入端口号，不输入直接回车则使用默认的 `8899` 端口
  - 当出现 `Server started ...` 信息时表示服务启动成功，此时不要关闭命令行窗口
  - 打开切片软件，设置物理打印机，输入命令行窗口中提示的 `http://127.0.0.1:端口号`，测试连接成功即可

```bash
## 自动查找模式
$ sm2uploader /path/to/code-file1 /path/to/code-file2
Discovering ...
Use the arrow keys to navigate: ↓ ↑ → ←
? Found 3 machines:
  ▸ A350-3DP@192.168.1.20 - Snapmaker A350
    A250-CNC@192.168.1.18 - Snapmaker A250
    J1V19@192.168.1.19 - Snapmaker-J1
Printer IP: 192.168.1.19
Printer Model: Snapmaker J1
Uploading file 'code-file1' [1.2 MB]...
  - SACP sending 100%
Upload finished.
Uploading file 'code-file2' [1.0 MB]...
  - SACP sending 100%
Upload finished.

## 指定打印机名字进行连接
$ sm2uploader -host J1V19 /path/to/code-file1
Discovering ...
Printer IP: 192.168.1.19
Printer Model: Snapmaker J1
Uploading file 'code-file1' [1.2 MB]...
  - SACP sending 100%
Upload finished.

## 模拟 OctoPrint (按下 CTRL+C 停止服务器)
$ sm2uploader -octoprint 127.0.0.1:8844 -host A350
Printer IP: 192.168.1.20
Printer Model: Snapmaker 2 Model A350
Starting OctoPrint server on :8844 ...
Server started, now you can upload files to http://127.0.0.1:8844
Request GET /api/version completed in 6.334µs
  - HTTP sending 100.0%
Upload finished: model.gcode [382.2 KB]
Request POST /api/files/local completed in 951.080458ms
```

打印机的 UDP 应答服务有时会挂掉，通常需要重启打印机来解决。或者你可以直接指定目标IP: `sm2uploader -host 192.168.1.20 /file.gcode`

如果 `host` 被发现过或者连接过，它会存在于 `knownhosts` 中，直接使用 id 进行连接会更加简洁: `sm2uploader -host A350-3DP /file.gcode`

更多参数：`sm2uploader -h`

## 在 macOS 系统提示文件无法打开的解决方法
macOS 不允许直接打开未经数字签名的程序，参考解决方案: https://osxdaily.com/2012/07/27/app-cant-be-opened-because-it-is-from-an-unidentified-developer/

也可以直接在终端执行 `xattr -d com.apple.quarantine sm2uploader-darwin`
