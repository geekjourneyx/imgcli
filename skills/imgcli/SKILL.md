---
name: imgcli
description: "Use this skill to run imgcli for agent-native image processing tasks: image inspection, normalization and format conversion, fixed-layout card composition, recipe-driven workflow execution, multi-platform variant export, smart padding to platform presets, packaging images into PDFs with visible text watermarks, and vertically stitching images into long composites with machine-readable JSON output."
---

# imgcli

用于指导 Agent 直接执行 `imgcli`，而不是解释底层图像算法实现。

## 项目边界

- `imgcli` 不负责生成图片。
- 图片生成应由外部工具完成，例如 Gemini、GPT Image 或其他 provider CLI/API。
- `imgcli` 负责对已经存在的图片做检查、加工、排版、打包和导出。

## 执行前检查

1. 先检查是否已构建：`imgcli version` 或 `./bin/imgcli version`
2. 如果未安装但仓库存在，优先执行：
```bash
make build
./bin/imgcli version
```
3. 默认使用 JSON 输出，不要关闭 `--json`
4. 输入路径必须存在；输出扩展名必须与目标格式匹配

## 常用工作流

### Inspect

```bash
./bin/imgcli inspect \
  --input input.jpg \
  --hash \
  --color-stats
```

可选：
- `--input-dir ./images`
- `--limit 10`

### Compose

```bash
./bin/imgcli compose \
  --input input.jpg \
  --output poster.jpg \
  --width 1080 \
  --height 1440 \
  --layout poster \
  --title "Launch Day" \
  --subtitle "A fixed-layout creator card"
```

可选：
- `--logo brand.png`
- `--badge NEW`
- `--background-color '#f3efe7'`
- `--safe-area 64,64,64,64`

### Convert

```bash
./bin/imgcli convert \
  --input alpha.png \
  --output out.jpg \
  --flatten-background "#ffffff" \
  --max-width 1600 \
  --quality 82 \
  --strip-metadata
```

### Run

```bash
./bin/imgcli run \
  --recipe recipe.json \
  --dry-run
./bin/imgcli run \
  --recipe recipe.json
```

引用规则：
- `input:<name>` 引用顶层输入
- `step:<id>` 引用前置 step 产出的文件

### Variants

```bash
./bin/imgcli variants \
  --input poster.jpg \
  --output-dir dist \
  --preset-set creator-basic
```

可选：
- `--preset xiaohongshu --preset wechat_cover`
- `--background blur|solid`
- `--filename-template '{base}_{preset}{ext}'`

### SmartPad

```bash
./bin/imgcli smartpad \
  --input input.jpg \
  --output output.jpg \
  --preset xiaohongshu
```

可选：
- `--background blur|solid`
- `--blur-sigma 5`
- `--quality 85`

### Images to PDF

```bash
./bin/imgcli topdf \
  --input page1.jpg \
  --input page2.jpg \
  --output bundle.pdf \
  --watermark-text "internal" \
  --watermark-position tile
```

可选：
- `--input-dir ./images`
- `--watermark-opacity 0.25`
- `--watermark-size 42`

### Vertical Stitch

```bash
./bin/imgcli stitch \
  --input a.jpg \
  --input b.jpg \
  --output stitched.jpg \
  --width 1080
```

可选：
- `--input-dir ./images`
- `--part-height-limit 65535`
- `--quality 85`

## 真实烟测

如果本机已经配置 `baoyu-image-gen`，执行：

```bash
make real-smoke
```

该流程会：
- 生成 3 张真实测试图
- 跑通 `inspect`
- 跑通 `compose`
- 跑通 `convert`
- 跑通 `variants`
- 跑通 `run`
- 跑通 `smartpad`、`topdf`、`stitch`
- 校验输出文件类型并打印产物目录

## 输出契约

成功：

```json
{"ok":true,"command":"smartpad","data":{}}
```

失败：

```json
{"error":"reason","code":"INVALID_ARGUMENT","exit_code":2}
```

## 失败时最小处理

- `PRESET_NOT_FOUND`：检查 `--preset`
- `INVALID_ARGUMENT`：检查必填 flag 与输入模式冲突
- `CONFIG_ERROR`：检查 `--recipe` 文件扩展名和 JSON/YAML 语法
- `PLAN_INVALID`：检查 step 引用、重复 step id、必填字段
- `OUTPUT_CONFLICT`：检查 recipe 是否把多个 step 写到同一路径，或覆盖顶层输入
- `DECODE_FAILED`：检查图片是否损坏或格式不支持
- `PDF_WRITE_FAILED`：检查输出路径是否可写
- `CANVAS_TOO_LARGE`：减小 `--width` 或调低分卷阈值
