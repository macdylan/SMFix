# Snapmaker2Slic3rPostProcessor
A Snapmaker2 G-Code Post Processor for PrusaSlicer and SuperSlicer to create compatible files for Snapmaker Touchscreen.

这里提供了一个针对 PrusaSlicer 切片时，修复 Gcode 文件头的小工具。

- Optimized gcode for printing on Snapmaker 2 (no scanning)
- Model thumbnails are displayed on touchscreen
- Support for multiple platforms including win/macOS/Linux

<br/>

- 优化 Gcode 文件，写入必要的元信息，避免打印机扫描整个文件（可能卡死）
- 在打印机触摸屏显示出模型的图片
- 支持平台 Win/macOS/Linux

## Install
- Download [smfix-{platform}-{arch}](https://github.com/macdylan/Snapmaker2Slic3rPostProcessor/releases/tag/go1.0)
    - Linux/macOS: `chmod +x smfix`
- PrusaSlicer Settings：
    1. `Printer Settings - Firmware - G-code thumbnails: 220x124`
    2. `Print Settings - Output options - Post-processing scripts: /path/to/smfix`
- Slice and export gcode file
- You can use [SM2Uploader](https://github.com/macdylan/sm2uploader) to quickly upload files to the printer

中文安装说明：

- [下载](https://github.com/macdylan/Snapmaker2Slic3rPostProcessor/releases/tag/go1.0) 适用于你设备的文件
    - Linux/macOS 下可能需要赋予可执行权限: `chmod +x smfix`
- PrusaSlicer 设置参数：
    1. 打印机设置 - 固件 - Gcode缩略图：220x124
    2. 打印设置 - 输出选项 - Post-processing：/path/to/smfix （建议路径不要有空格）
- 设置完成，切片并导出 Gcode 即可
- 如果希望更快的完成文件上传，可以使用 [SM2Uploader](https://github.com/macdylan/sm2uploader)，参考[快速设置说明](https://github.com/macdylan/sm2uploader/wiki)。
