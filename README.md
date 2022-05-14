# Snapmaker2Slic3rPostProcessor
A Snapmaker2 G-Code Post Processor for PrusaSlicer and SuperSlicer to create compatible files for Snapmaker Touchscreen.

- Optimized gcode for printing on Snapmaker 2 (no scanning)
- Model thumbnails are displayed on touchscreen
- Support for multiple platforms including win/macOS/Linux

## Install
- Download [smfix-{platform}-{arch}](https://github.com/macdylan/Snapmaker2Slic3rPostProcessor/releases/tag/go1.0)
    - Linux/macOS: `chmod +x smfix`
- PrusaSlicer Settings：
    1. `Printer Settings - Firmware - G-code thumbnails: 220x124`
    2. `Print Settings - Output options - Post-processing scripts: /path/to/smfix`
- Slice and export gcode file
- You can use [SM2Uploader](https://github.com/macdylan/sm2uploader) to quickly upload files to the printer
