A Snapmaker2 G-Code Post Processor for PrusaSlicer/SuperSlicer/OrcaSlicer to create compatible files for Snapmaker printers.

- Optimized gcode for printing on Snapmaker 2 (no scanning)
- Model thumbnails are displayed on touchscreen
- Smart pre-heat for switch tools, shutoff nozzles that are no longer in use, and other optimization features for multi-extruders.
- Support for multiple platforms including win/macOS/Linux

## Install
- Download [Latest version](https://github.com/macdylan/Snapmaker2Slic3rPostProcessor/releases)
    - Linux/macOS: `chmod +x smfix`
    - macOS prohibits opening unsigned programs. Please refer to the [solution](https://osxdaily.com/2012/07/27/app-cant-be-opened-because-it-is-from-an-unidentified-developer/):
        - Remove the restriction by execute in terminal: `xattr -d com.apple.quarantine /path/to/smfix-darwin`

- PrusaSlicer Settings：
    1. `Printer Settings - Firmware - G-code thumbnails: 220x124`
    2. `Print Settings - Output options - Post-processing scripts: /path/to/smfix`
- Slice and export gcode file
- Alternatively, you can also use this tool separately in the Terminal. For specific instructions, please refer to the information provided in the `-h` flag.

If you are unable to utilize advanced features such as smart pre-heat or disabling inactive nozzles on your multi-extruders(J1 / Dual-extruder Module), it is possible that there is something mistakes in your slicer settings. You can directly use [my configuration parameters](https://github.com/macdylan/3dp-configs) as an alternative.

## About sm2uploader:
Since [sm2uploader v2.0](https://github.com/macdylan/sm2uploader/releases), all the functionalities of SMFix have been integrated, allowing for a seamless repair and network upload. The main purpose of retaining SMFix is to cater to scenarios where print using a USB drive.

Please refer to the [Quick Setup Guide](https://github.com/macdylan/sm2uploader/wiki) for further instructions.

----

## Chinese
这是一个在 PrusaSlicer/SuperSlicer/OrcaSlicer 切片时，修复 Gcode 文件头的小工具。

- 优化 Gcode 文件，写入必要的元信息，避免打印机扫描整个文件以至于失去响应
- 在打印机屏幕显示出模型的图片
- 为多挤出机提供智能预热、关闭不再使用的喷头等优化功能
- 支持平台 Win/macOS/Linux

## 安装说明：
- [下载](https://github.com/macdylan/SMFix/releases) 适用于你设备的文件
    - Linux/macOS 下可能需要赋予可执行权限: `chmod +x smfix`
    - macOS 不允许打开未经数字签名的程序，参考[解决方案](https://osxdaily.com/2012/07/27/app-cant-be-opened-because-it-is-from-an-unidentified-developer/):
        - 去除限制，在终端执行 `xattr -d com.apple.quarantine /path/to/smfix-darwin`

- PrusaSlicer 设置参数：
    1. 打印机设置 - 固件 - Gcode缩略图：220x124
    2. 打印设置 - 输出选项 - Post-processing：/path/to/smfix （建议路径不要有空格）
- 设置完成，切片并导出 Gcode 即可
- 也可以在终端里单独使用这个工具，具体请参考 `-h` 的说明信息

在多挤出机上使用本工具，如果无法实现智能预热、关闭停用的喷头等高级功能，可能是你的切片软件设置错误。你可以直接使用[我的配置参数](https://github.com/macdylan/3dp-configs)。

## 关于 sm2uploader
从 [sm2uploader v2.0](https://github.com/macdylan/sm2uploader/releases) 开始，已经集成了 SMFix 的所有功能，可一步完成修复和网络上传的功能。保留 SMFix 的主要目的是为了使用 U 盘进行打印的场景。

参考[快速设置说明](https://github.com/macdylan/sm2uploader/wiki)。
