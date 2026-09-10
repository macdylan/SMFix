# SMFix 测试用例清单

> 从 REQUIREMENTS.md 逐条映射。状态标记：✅ 已有测试 | ❌ 缺失 | ⚠️ 部分覆盖

---

## §2 打印机型号识别

### 2.0 通过 printer_model 识别

| # | 用例 | 输入 | 期望 | 状态 |
|---|------|------|------|------|
| 2.0.1 | A150 | `printer_model = SNAPMAKER A150` | Model = `Snapmaker 2.0 A150` | ✅ ModelFromPrinterModel |
| 2.0.2 | A250 | `printer_model = SNAPMAKER A250` | Model = `Snapmaker 2.0 A250` | ✅ ModelFromPrinterModel |
| 2.0.3 | A350 | `printer_model = SNAPMAKER A350` | Model = `Snapmaker 2.0 A350` | ✅ ModelFromPrinterModel |
| 2.0.4 | Artisan/A400 | `printer_model = SNAPMAKER Artisan` | Model = `A400` | ✅ ModelFromPrinterModel |
| 2.0.5 | J1 | `printer_model = SNAPMAKER J1` | Model = `Snapmaker J1`, Version = 1 | ✅ ModelFromPrinterModel |
| 2.0.6 | U1 | `printer_model = SNAPMAKER U1` | Model = `Snapmaker U1` | ✅ ModelFromPrinterModel |

### 2.0.B 通过 bed_shape 识别

| # | 用例 | 输入 | 期望 | 状态 |
|---|------|------|------|------|
| 2.0.B.1 | A150 (160x160) | `bed_shape = 0x0,160x160` | Model = `Snapmaker 2.0 A150` | ✅ ModelFromBedShape |
| 2.0.B.2 | A250 (230x250) | `bed_shape = 0x0,230x250` | Model = `Snapmaker 2.0 A250` | ✅ ModelFromBedShape |
| 2.0.B.3 | A250 dual+qskit (220x235) | `bed_shape = 0x0,220x235` | Model = `Snapmaker 2.0 A250` | ❌ 缺失 |
| 2.0.B.4 | A350 (320x350) | `bed_shape = 0x0,320x350` | Model = `Snapmaker 2.0 A350` | ✅ ModelFromBedShape |
| 2.0.B.5 | A350 dual (310x350) | `bed_shape = 0x0,310x350` | Model = `Snapmaker 2.0 A350` | ❌ 缺失 |
| 2.0.B.6 | A350 qskit (320x335) | `bed_shape = 0x0,320x335` | Model = `Snapmaker 2.0 A350` | ❌ 缺失 |
| 2.0.B.7 | A350 dual+qskit (310x335) | `bed_shape = 0x0,310x335` | Model = `Snapmaker 2.0 A350` | ❌ 缺失 |
| 2.0.B.8 | A400 (400x400) | `bed_shape = 0x0,400x400` | Model = `A400` | ✅ ModelFromBedShape |
| 2.0.B.9 | J1 (312x200) | `bed_shape = 0x0,312x200` | Model = `Snapmaker J1`, Version = 1 | ✅ ModelFromBedShape |
| 2.0.B.10 | J1 (324x200) | `bed_shape = 0x0,324x200` | Model = `Snapmaker J1`, Version = 1 | ❌ 缺失 |
| 2.0.B.11 | J1 (300x200) | `bed_shape = 0x0,300x200` | Model = `Snapmaker J1`, Version = 1 | ❌ 缺失 |
| 2.0.B.12 | U1 (270x270) | `bed_shape = 0x0,270x270` | Model = `Snapmaker U1` | ✅ ModelFromBedShape |

### 2.1 挤出头类型

| # | 用例 | 输入 | 期望 | 状态 |
|---|------|------|------|------|
| 2.1.1 | 默认单挤出头 | 仅 T0 有耗材用量 | ToolHead = single | ✅ SingleExtruderUnused |
| 2.1.2 | 双挤出头（耗材用量） | T0 和 T1 均有耗材用量 | ToolHead = dual | ✅ ValidPrusaSlicer |
| 2.1.3 | 双挤出头（model 含 Dual） | `printer_model = A350 Dual` | ToolHead = dual | ✅ DualFromModelName |
| 2.1.4 | 双挤出头（notes 含 _DUAL） | `printer_notes = ..._DUAL...` | ToolHead = dual | ✅ DualFromPrinterNotes |
| 2.1.5 | 仅 T1 使用 | T0 耗材 = 0，T1 耗材 > 0 | LeftUsed=false, RightUsed=true, ToolHead=single | ❌ 缺失 |

### 2.2 打印模式

| # | 用例 | 输入 | 期望 | 状态 |
|---|------|------|------|------|
| 2.2.1 | Default | 无 M605 | PrintMode = Default | ✅ ValidPrusaSlicer |
| 2.2.2 | Duplication | `M605 S2` | PrintMode=Duplication, Model=J1, Version=1 | ✅ PrintModes |
| 2.2.3 | Mirror | `M605 S3` | PrintMode=Mirror, Model=J1, Version=1 | ✅ PrintModes |
| 2.2.4 | Backup | `M605 S4` | PrintMode=Backup | ✅ PrintModes |

---

## §3 输入/输出规格

### 3.1 输入

| # | 用例 | 输入 | 期望 | 状态 |
|---|------|------|------|------|
| 3.1.1 | 幂等性拒绝 | 首行 `; Postprocessed by smfix` | 返回 ErrIsFixed | ✅ ErrIsFixed |
| 3.1.2 | 空文件 | 0 行 | 返回 ErrInvalidGcode | ❌ 缺失 |
| 3.1.3 | 纯注释文件 | 仅 `;` 注释行 | 返回 ErrInvalidGcode | ❌ 缺失 |

### 3.3 有效性校验

| # | 用例 | 输入 | 期望 | 状态 |
|---|------|------|------|------|
| 3.3.1 | 行数不足 | < 20 行 | ErrInvalidGcode | ✅ ErrInvalidGcode_TooFewLines |
| 3.3.2 | 无型号 | 有温度，无 printer_model/bed_shape | ErrInvalidGcode | ✅ ErrInvalidGcode_NoModel |
| 3.3.3 | 无温度 | 有型号，无 first_layer_temperature | ErrInvalidGcode | ✅ ErrInvalidGcode_NoTemperature |
| 3.3.4 | 恰好 20 行 | 刚好 20 行 + 有效参数 | 成功 | ❌ 缺失（边界值） |
| 3.3.5 | 19 行 | 19 行 + 有效参数 | ErrInvalidGcode | ❌ 缺失（边界值） |

---

## §4 元数据解析

### 4.1 PrusaSlicer 格式

| # | 用例 | 参数键 | 期望 | 状态 |
|---|------|--------|------|------|
| 4.1.1 | 层高 | `layer_height = 0.2` | LayerHeight = 0.2 | ✅ ValidPrusaSlicer |
| 4.1.2 | 总层数 | `total_layer_number = 150` | TotalLayers = 150 | ✅ ValidPrusaSlicer |
| 4.1.3 | 喷嘴温度 | `first_layer_temperature = 215, 220` | NozzleTemps = [215, 220] | ✅ ValidPrusaSlicer |
| 4.1.4 | 热床温度 | `first_layer_bed_temperature = 60, 65` | BedTemps = [60, 65] | ✅ ValidPrusaSlicer |
| 4.1.5 | 回抽长度 | `retract_length = 1.2, 1.5` | Retractions = [1.2, 1.5] | ✅ RetractionFallback |
| 4.1.6 | 打印速度 | `max_print_speed = 100` | PrintSpeedSec = 100 | ✅ ValidPrusaSlicer |
| 4.1.7 | 预估时间 | `estimated printing time = 1h 30m 0s` | EstimatedTimeSec = 5400 | ✅ ValidPrusaSlicer |

### 4.1.B Bambu Studio / OrcaSlicer 格式

| # | 用例 | 参数键 | 期望 | 状态 |
|---|------|--------|------|------|
| 4.1.B.1 | 层高 | `first_layer_height = 0.2` | LayerHeight = 0.2 | ✅ ValidBBS |
| 4.1.B.2 | 总层数 | `total layers count = 200` | TotalLayers = 200 | ✅ ValidBBS |
| 4.1.B.3 | 喷嘴温度 | `nozzle_temperature_initial_layer = 220, 230` | NozzleTemps = [220, 230] | ✅ ValidBBS |
| 4.1.B.4 | 热床温度 | `hot_plate_temp_initial_layer = 55, 60` | BedTemps = [55, 60] | ✅ ValidBBS |
| 4.1.B.5 | 回抽长度 | `filament_retraction_length = 1.0, 1.2` | Retractions = [1.0, 1.2] | ✅ ValidBBS |
| 4.1.B.6 | 打印速度 | `outer_wall_speed = 80` | PrintSpeedSec = 80 | ✅ ValidBBS |
| 4.1.B.7 | 耗材类型分号分隔 | `filament_type = PETG;ASA` | FilamentTypes = [PETG, ASA] | ✅ ValidBBS |

### 4.2 完整参数列表

| # | 用例 | 参数 | 期望 | 状态 |
|---|------|------|------|------|
| 4.2.1 | 喷嘴直径 | `nozzle_diameter = 0.4, 0.4` | NozzleDiameters = [0.4, 0.4] | ✅ ValidPrusaSlicer |
| 4.2.2 | 耗材类型 | `filament_type = PLA,ABS` | FilamentTypes = [PLA, ABS] | ✅ ValidPrusaSlicer |
| 4.2.3 | 耗材用量 mm | `filament used [mm] = 5000, 3000` | FilamentUsed = [5000, 3000] | ✅ ValidPrusaSlicer |
| 4.2.4 | 耗材重量 g | `filament used [g] = 12.5, 7.5` | FilamentUsedWeight = [12.5, 7.5] | ✅ ValidPrusaSlicer |
| 4.2.5 | 换工具回抽 | `retract_length_toolchange = 3.0, 3.5` | SwitchRetraction = [3.0, 3.5] | ✅ ValidPrusaSlicer |
| 4.2.6 | Min XYZ | `min_x/y/z = 10/20/0.2` | MinX=10, MinY=20, MinZ=0.2 | ✅ ValidPrusaSlicer |
| 4.2.7 | Max XYZ | `max_x/y/z = 300/330/50.4` | MaxX=300, MaxY=330, MaxZ=50.4 | ✅ ValidPrusaSlicer |
| 4.2.8 | printer_notes | `printer_notes = ...` | PrinterNotes 正确存储 | ⚠️ 间接覆盖 |

### 4.3 挤出头使用状态

| # | 用例 | 输入 | 期望 | 状态 |
|---|------|------|------|------|
| 4.3.1 | T0 未使用 | `filament used [mm] = 0, 3000` | T0 温度=0, 类型=`-`, 回抽=0, 热床=-1 | ✅ SingleExtruderUnused |
| 4.3.2 | T1 未使用 | `filament used [mm] = 5000, 0` | T1 温度=0, 类型=`-`, 回抽=0, 热床=-1 | ❌ 缺失 |
| 4.3.3 | 两者均未使用 | `filament used [mm] = 0, 0` | 两者均重置 | ❌ 缺失 |

### 4.4 回抽优先级

| # | 用例 | 输入 | 期望 | 状态 |
|---|------|------|------|------|
| 4.4.1 | filament_retract_length 覆盖 | 同时有 retract_length 和 filament_retract_length | 取 filament_retract_length | ✅ RetractionFallback |
| 4.4.2 | 仅有 retract_length | 只有 retract_length | 取 retract_length | ✅ RetractionFallback |
| 4.4.3 | 两者均无 | 无回抽参数 | Retractions 保持初始值 | ❌ 缺失 |
| 4.4.4 | 仅 T0 有 filament_retract | `filament_retract_length = 1.0, 0` + `retract_length = 0.5, 0.6` | T0=1.0(覆盖), T1=0.6(原始) | ❌ 缺失 |

### 4.5 版本判定

| # | 用例 | 输入 | 期望 | 状态 |
|---|------|------|------|------|
| 4.5.1 | 默认 V0 | A 系列打印机 | Version = 0 | ✅ ValidPrusaSlicer |
| 4.5.2 | J1 强制 V1 | J1 打印机 | Version = 1 | ✅ ValidBBS |
| 4.5.3 | 头部标记 V1 | `; SNAPMAKER_GCODE_V1` | Version = 1 | ✅ SNAPMAKER_GCODE_V1_Header |
| 4.5.4 | notes 覆盖为 V1 | `printer_notes = SNAPMAKER_GCODE_V1 ...` | Version = 1 | ✅ VersionFromPrinterNotes |
| 4.5.5 | notes 覆盖为 V0 | `printer_notes = SNAPMAKER_GCODE_V0 ...` + 头部 V1 | Version = 0（notes 优先） | ✅ VersionFromPrinterNotes |
| 4.5.6 | IDEX Duplication 强制 V1 | `M605 S2` | Version = 1 | ✅ PrintModes |

### 4.6 总行数计算

| # | 用例 | 输入 | 期望 | 状态 |
|---|------|------|------|------|
| 4.6.1 | `; generated by` 重置计数 | 前面有 3 行，然后 `; generated by` | TotalLines 从 1 重新计数 | ✅ TotalLinesReset |
| 4.6.2 | 无 `; generated by` | 直接从第一行计数 | TotalLines = 实际行数 | ❌ 缺失 |

### 4.7 有效温度

| # | 用例 | 输入 | 期望 | 状态 |
|---|------|------|------|------|
| 4.7.1 | T0 优先 | T0=215, T1=220 | EffectiveNozzle = 215 | ✅ EffectiveTemperatures |
| 4.7.2 | T0 < 1 回退 T1 | T0=0, T1=220 | EffectiveNozzle = 220 | ⚠️ 测试存在但未断言返回值 |
| 4.7.3 | 有效热床温度 T0 优先 | bed0=60, bed1=65 | EffectiveBed = 60 | ✅ EffectiveTemperatures |

---

## §5 头部生成

### 5.1 V0 格式

| # | 用例 | 期望 | 状态 |
|---|------|------|------|
| 5.1.1 | V0 头部包含 Mark 行 | 首行 = `; Postprocessed by smfix ...` | ❌ 缺失 |
| 5.1.2 | V0 machine 字段 | `;machine: Snapmaker 2.0 A350` | ❌ 缺失 |
| 5.1.3 | V0 tool_head 字段 | `;tool_head: singleExtruderToolheadForSM2` 或 dual | ❌ 缺失 |
| 5.1.4 | V0 file_total_lines = TotalLines + 34 | 行数正确偏移 | ❌ 缺失 |
| 5.1.5 | V0 estimated_time × 1.07 | 时间乘以系数 | ❌ 缺失 |
| 5.1.6 | V0 work_speed × 60 | mm/s → mm/min | ❌ 缺失 |
| 5.1.7 | V0 含缩略图 | `;thumbnail: data:image/png;base64,...` | ❌ 缺失 |
| 5.1.8 | V0 无缩略图 | 无 thumbnail 行 | ❌ 缺失 |
| 5.1.9 | V0 所有温度/直径/材料/回抽字段 | 值正确 | ❌ 缺失 |

### 5.2 V1 格式

| # | 用例 | 期望 | 状态 |
|---|------|------|------|
| 5.2.1 | V1 Version 字段 | `;Version:1` | ❌ 缺失 |
| 5.2.2 | V1 Printer 字段 | `;Printer:Snapmaker J1` | ❌ 缺失 |
| 5.2.3 | V1 Lines = TotalLines + 27 | 行数正确偏移 | ❌ 缺失 |
| 5.2.4 | V1 Extruder Mode | `;Extruder Mode:Default` 等 | ❌ 缺失 |
| 5.2.5 | V1 Extruder(s) Used = 1 | 单挤出头 | ❌ 缺失 |
| 5.2.6 | V1 Extruder(s) Used = 2 | 双挤出头 | ❌ 缺失 |
| 5.2.7 | V1 含缩略图 | `;Thumbnail:data:image/...` | ❌ 缺失 |
| 5.2.8 | V1 无缩略图 | 无 Thumbnail 行 | ❌ 缺失 |

---

## §6 G-code 优化功能

### 6.1 智能关闭喷嘴 (GcodeFixShutoff)

| # | 用例 | 输入 | 期望 | 状态 |
|---|------|------|------|------|
| 6.1.1 | 切换后关闭旧挤出头 | T0 → T1 | 插入 `M104 S0 T0` | ✅ GcodeShutoff |
| 6.1.2 | 关闭后温度命令被注释 | T1 关闭后出现 M104 T1 | 注释为 `has been shutted off` | ✅ GcodeShutoff |
| 6.1.3 | 关闭后 M109 也被注释 | T1 关闭后出现 M109 T1 | 注释为 `has been shutted off` | ✅ GcodeShutoff |
| 6.1.4 | 最终 S0 命令保留 | 末尾 `M104 S0 T1` | 保持原样不注释 | ✅ GcodeShutoff |
| 6.1.5 | 无切换不关闭 | 仅 T0 使用 | 无 shutoff 插入 | ❌ 缺失 |
| 6.1.6 | 多次切换同一工具 | T0→T1→T0→T1 | 每次切换正确跟踪 | ❌ 缺失 |

### 6.2 智能预热 (GcodeFixPreheat)

| # | 用例 | 输入 | 期望 | 状态 |
|---|------|------|------|------|
| 6.2.1 | 深冻 | standby 后 ≥3 次 M73 变化 | 替换为 110°C + 插入预热 | ✅ GcodePreheat |
| 6.2.2 | 短预热 | standby 后 1-2 次 M73 变化 | 插入 M104 预热 | ✅ GcodePreheat |
| 6.2.3 | 长预热 | standby 后 ≥3 次 M73 变化 | 在更早位置插入预热 | ✅ GcodePreheat |
| 6.2.4 | 冷却移除 | standby 后 <1 次 M73 变化 | 删除 standby 命令 | ✅ GcodePreheat |
| 6.2.5 | 重复温度请求 | 同一工具同温 M104 重复 | 注释为 `already requested` | ✅ GcodePreheat |
| 6.2.6 | 重复温度等待 | 同一工具同温 M109 重复 | 首次保留，后续注释 `already stabilized` | ✅ GcodePreheat |
| 6.2.7 | 无 M73 不修改 | 有 M109 但无 M73 | 不做任何修改 | ❌ 缺失 |
| 6.2.8 | 不同温度保留 | M109 温度不同 | 保留原命令 | ✅ GcodePreheat (隐含) |

### 6.3 工具号重映射 (GcodeReplaceToolNum)

| # | 用例 | 输入 | 期望 | 状态 |
|---|------|------|------|------|
| 6.3.1 | T0 不变 | `T0` | `T0` | ✅ ReplaceToolNum |
| 6.3.2 | T1 不变 | `T1` | `T1` | ✅ ReplaceToolNum |
| 6.3.3 | T2→T0 | `T2` | `T0` | ✅ ReplaceToolNum |
| 6.3.4 | T3→T1 | `T3` | `T1` | ✅ ReplaceToolNum |
| 6.3.5 | T63→T1 | `T63` | `T1` | ✅ ReplaceToolNum |
| 6.3.6 | M104 T 参数重映射 | `M104 T3 S200` | `M104 T1 S200` | ✅ ReplaceToolNum |
| 6.3.7 | M106 P 参数重映射 | `M106 P62 S200` | `M106 P0 S200` | ✅ ReplaceToolNum |
| 6.3.8 | M107 P 参数重映射 | `M107 P62` | `M107 P0` | ✅ ReplaceToolNum |
| 6.3.9 | M301 E 参数重映射 | `M301 E3 T0` | `M301 E1 T0` | ✅ ReplaceToolNum |
| 6.3.10 | M303 E 参数重映射 | `M303 E4 T1` | `M303 E0 T1` | ✅ ReplaceToolNum |
| 6.3.11 | 非相关命令不修改 | `M108 T33` | `M108 T33`（不变） | ✅ ReplaceToolNum |
| 6.3.12 | G 命令 T 参数不修改 | `G1 T33` | `G1 T33`（不变） | ✅ ReplaceToolNum |
| 6.3.13 | 注释参数重排 (2 工具) | `filament used = a, b` | 按 idxT0/idxT1 重排 | ✅ ReplaceComment |
| 6.3.14 | 注释参数重排 (4→2 工具) | T3/T2 + 4 值注释 | 取 idx 对应值 | ✅ ReplaceComment |
| 6.3.15 | 注释分号分隔 | `filament_type = PLA;ABS; PETG; ASA` | 正确处理分号分隔 | ✅ ReplaceComment |
| 6.3.16 | 无 data race | 并行处理 | `-race` 检测通过 | ✅ go test -race |

### 6.4 Orca 换料清理 (GcodeFixOrcaToolUnload)

| # | 用例 | 输入 | 期望 | 状态 |
|---|------|------|------|------|
| 6.4.1 | 移除无效 M104 | TOOLCHANGE 区间内 `M104 S278`（无 T） | 注释为 `remove` | ✅ OrcaToolUnload |
| 6.4.2 | 保留有效 M104 | TOOLCHANGE 区间内 `M104 S250 T0` | 保持原样 | ✅ OrcaToolUnload |
| 6.4.3 | 注释中 T 不算 | `M104 S255; T1`（T 在注释中） | GetToolNum 从注释提取到 T，保留 | ✅ OrcaToolUnload (隐含) |
| 6.4.4 | 非 TOOLCHANGE 区间不处理 | TOOLCHANGE 外的 M104 无 T | 保持原样 | ❌ 缺失 |
| 6.4.5 | 空输入 | 无 gcode | 返回空 | ❌ 缺失 |

### 6.6 预处理

| # | 用例 | 输入 | 期望 | 状态 |
|---|------|------|------|------|
| 6.6.1 | G4 S0 过滤 | `G4 S0` | 被跳过不进入 gcodes 列表 | ❌ 缺失（smfix.go 层） |
| 6.6.2 | G4 P100 保留 | `G4 P100` | 正常保留 | ❌ 缺失（smfix.go 层） |
| 6.6.3 | 前导空格裁剪 | ` G0 X1` | `G0 X1` | ✅ ParseGcodeBlock (间接) |
| 6.6.4 | 连续空格合并 | `G0  X1   Y2` | `G0 X1 Y2` | ✅ ParseGcodeBlock (间接) |
| 6.6.5 | 特殊字符移除 | `G0\n X1` | `G0 X1` | ✅ prepareGcodeLineToParse (间接) |
| 6.6.6 | 尾部空格裁剪 | `G0 X1 ` | `G0 X1` | ✅ ParseGcodeBlock (间接) |
| 6.6.7 | 空行 | 空字符串 | 返回 ErrEmptyString | ✅ ParseGcodeBlock (间接) |
| 6.6.8 | 纯注释行 | `; comment` | 仅 comment，无 cmd | ✅ ParseGcodeBlock (间接) |

---

## §7 缩略图处理

| # | 用例 | 输入 | 期望 | 状态 |
|---|------|------|------|------|
| 7.1 | 多缩略图取最后一个 | 16x16 和 220x124 两个 | 取 220x124 的数据 | ✅ ConvertThumbnail |
| 7.2 | base64 格式 | thumbnail 数据 | `data:image/png;base64,...` | ✅ ConvertThumbnail |
| 7.3 | 去除注释前缀 | `; AAAA` → `AAAA` | 分号+空格被移除 | ✅ ConvertThumbnail |
| 7.4 | 通过 ParseParams 集成 | 完整 gcode 含 thumbnail | Params.Thumbnail 正确 | ✅ ParseParams_Thumbnail |
| 7.5 | 无缩略图 | 无 thumbnail begin/end | Params.Thumbnail 为空 | ❌ 缺失 |
| 7.6 | 空缩略图 | thumbnail begin 后直接 end | 处理不崩溃 | ❌ 缺失 |

---

## §8 命令行接口（集成测试）

| # | 用例 | 期望 | 状态 |
|---|------|------|------|
| 8.1 | 无参数显示帮助 | 打印 usage 并退出 | ❌ 缺失 |
| 8.2 | 正常处理文件 | 输出文件含头部 + 优化后 gcode | ❌ 缺失 |
| 8.3 | `-o` 指定输出 | 写入指定路径 | ❌ 缺失 |
| 8.4 | `-noshutoff` | 不执行 shutoff 优化 | ❌ 缺失 |
| 8.5 | `-noreplacetool` | 不执行工具号重映射 | ❌ 缺失 |
| 8.6 | 已处理文件 | 报错 `No need to fix again` | ❌ 缺失 |
| 8.7 | 无效文件 | 报错退出 | ❌ 缺失 |
| 8.8 | 不存在的输入文件 | 报错退出 | ❌ 缺失 |

---

## 汇总

| 分类 | ✅ 已有 | ⚠️ 部分覆盖 | ❌ 缺失 | 总计 |
|------|---------|-------------|---------|------|
| §2 打印机识别 | 14 | 0 | 6 | 20 |
| §3 输入/输出 | 3 | 0 | 5 | 8 |
| §4 元数据解析 | 28 | 2 | 8 | 38 |
| §5 头部生成 | 0 | 0 | 17 | 17 |
| §6 G-code 优化 | 28 | 0 | 9 | 37 |
| §7 缩略图 | 4 | 0 | 2 | 6 |
| §8 CLI 集成 | 0 | 0 | 8 | 8 |
| **合计** | **77** | **2** | **55** | **134** |

### 优先级建议

**P0 — 高价值缺失（核心功能无覆盖）**
1. §5 头部生成（17 个用例全部缺失，header.go 0% 覆盖）
2. §6.2.7 无 M73 时预热不修改
3. §6.4.4 非 TOOLCHANGE 区间不处理

**P1 — 边界和异常路径**
4. §3.1.2/3.1.3 空文件/纯注释
5. §3.3.4/3.3.5 行数边界值
6. §4.3.2/4.3.3 T1 未使用/两者均未使用
7. §7.5/7.6 无缩略图/空缩略图

**P2 — 集成测试**
8. §8 CLI 端到端测试（8 个用例全部缺失）
