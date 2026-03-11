# imgcli

[![CI](https://github.com/geekjourneyx/imgcli/actions/workflows/ci.yml/badge.svg)](https://github.com/geekjourneyx/imgcli/actions/workflows/ci.yml)
[![Release](https://github.com/geekjourneyx/imgcli/actions/workflows/release.yml/badge.svg)](https://github.com/geekjourneyx/imgcli/actions/workflows/release.yml)
[![Go Version](https://img.shields.io/badge/go-1.26%2B-00ADD8?logo=go)](https://go.dev/)
[![CGO Free](https://img.shields.io/badge/cgo-free-2F855A)](https://go.dev/)

[简体中文](./README.zh-CN.md) | [English](./README.md)

`imgcli` 是一个 **纯 Go、面向 Agent、默认 JSON 输出** 的图像后处理 CLI。它面向脚本、Worker、数字员工和自动化流水线，目标不是“做一切图像能力”，而是把**已经存在的图片资产**稳定地检查、加工、排版、打包和导出。

## 项目原则

- `imgcli` 是图片**后处理与打包** CLI，不是图片生成 CLI。
- 图片生成应交给外部工具，例如 Gemini、GPT Image 或其他 provider CLI / API。
- `imgcli` 负责处理**已经存在**的图片：检查、规范化、排版、拼接、打包和导出。
- 核心能力必须保持确定性、本地文件优先、适合自动化和批处理。

## 为什么做 imgcli

很多图像工具不适合 Agent 和自动化，不是因为算法不够强，而是因为接口不稳定：

- 输出偏人类可读，不适合机器消费。
- 大图批处理时容易内存失控。
- 社媒封面常常靠粗暴裁剪，主体被切坏。
- PDF、水印、长图这些高频能力经常依赖 GUI 工具或重量级运行时。

`imgcli` 的定位是收窄边界，只把最适合自动化的部分做好：

- 默认 JSON 输出，方便程序消费
- 稳定错误码和非 0 退出码
- CGO-free，方便跨平台发布和部署
- 顺序处理、低内存占用、适合批量任务

## 当前核心命令

- `inspect`
  - 检查图片元数据，帮助 Agent 在处理前先“看懂图”
  - 可选输出颜色统计、SHA256、感知哈希
  - 支持单文件和目录扫描
- `compose`
  - 基于一张已有图片渲染固定布局卡片
  - 支持标题、副标题、Logo、角标、安全区和图片圆角
  - 严格限制在少量命名布局家族内，不做自由排版
- `convert`
  - 做格式转换、尺寸上限约束和 JPEG 交付规范化
  - 支持透明图转 JPEG 时的背景扁平化
  - 通过确定性的重编码路径去掉原始元数据
- `run`
  - 用 JSON 或 YAML recipe 执行多步处理，不通过 shell 回调子命令
  - 支持 `input:<name>` 和 `step:<id>` 两种显式引用
  - 支持 `--dry-run` 先查看解析后的执行计划
- `variants`
  - 基于一张源图一次导出多个平台版本
  - 支持 `creator-basic` 这类内置 preset set
  - 保持确定性的输出命名，并直接复用 `smartpad` 能力
- `smartpad`
  - 将图片适配到目标平台尺寸，例如 `xiaohongshu`、`wechat_cover`
  - 保持主体完整，使用留白和背景填充而不是粗暴裁切
  - 支持模糊背景和纯色背景
- `topdf`
  - 将多张图片按稳定顺序打包为 PDF
  - 支持在入 PDF 过程中叠加显性文字水印
  - 按页顺序处理，避免整批图片常驻内存
- `stitch`
  - 将多张图片按统一宽度自上而下拼接为长图
  - 超长时自动分卷输出
  - JSON 返回所有产物路径

## 安装

推荐安装方式：

```bash
curl -fsSL https://raw.githubusercontent.com/geekjourneyx/imgcli/main/scripts/install.sh | bash
```

安装脚本会：

- 自动识别 `linux` / `darwin`
- 下载正确的 release 产物
- 用 `SHA256SUMS` 做校验
- 安装到 `~/.local/bin`
- 必要时提示你补 PATH

手动下载地址：

- Releases：`https://github.com/geekjourneyx/imgcli/releases`
- 产物：
  - `imgcli-linux-amd64`
  - `imgcli-linux-arm64`
  - `imgcli-darwin-amd64`
  - `imgcli-darwin-arm64`
  - `SHA256SUMS`

验证安装：

```bash
imgcli version
```

## 快速开始

```bash
make build
./bin/imgcli version

./bin/imgcli inspect \
  --input in.jpg \
  --hash \
  --color-stats

./bin/imgcli compose \
  --input in.jpg \
  --output poster.jpg \
  --width 1080 \
  --height 1440 \
  --layout poster \
  --title "Launch Day" \
  --subtitle "固定布局创作者卡片"

./bin/imgcli convert \
  --input source.png \
  --output normalized.jpg \
  --flatten-background "#ffffff" \
  --max-width 1600 \
  --quality 82 \
  --strip-metadata

./bin/imgcli run \
  --recipe recipe.json \
  --dry-run

./bin/imgcli variants \
  --input poster.jpg \
  --output-dir dist \
  --preset-set creator-basic

./bin/imgcli smartpad \
  --input in.jpg \
  --output out.jpg \
  --preset xiaohongshu

./bin/imgcli topdf \
  --input page1.jpg \
  --input page2.jpg \
  --output bundle.pdf \
  --watermark-text "internal"

./bin/imgcli stitch \
  --input a.jpg \
  --input b.jpg \
  --output stitched.jpg \
  --width 1080
```

## JSON 输出契约

`imgcli` 默认输出 JSON。  
成功结果写到 `stdout`；失败时输出结构化 JSON 到 `stderr`，并返回非 0 退出码。

成功示例：

```json
{
  "ok": true,
  "command": "smartpad",
  "data": {
    "input": "in.jpg",
    "output": "out.jpg",
    "target_width": 1080,
    "target_height": 1440
  }
}
```

失败示例：

```json
{
  "error": "preset \"foo\" not found",
  "code": "PRESET_NOT_FOUND",
  "exit_code": 2
}
```

契约规则：

- JSON 字段默认只做增量新增，不随意破坏旧字段
- 调用方不能依赖帮助文案或文本输出来解析
- 错误处理必须以 `code` 为准，不要依赖错误字符串

## Recipe 引用规则

`run` 仍然坚持“文件路径优先”，只额外支持两种显式引用：

- `input:<name>`：引用 recipe 顶层 `inputs` 里的命名输入
- `step:<id>`：引用前一个 step 产出的文件输出

最小示例：

```json
{
  "version": "v1",
  "inputs": {
    "hero": "hero.jpg",
    "logo": "logo.png"
  },
  "steps": [
    {
      "id": "card",
      "type": "compose",
      "input": "input:hero",
      "output": "dist/card.jpg",
      "width": 1080,
      "height": 1440,
      "layout": "poster",
      "title": "Launch Day",
      "logo": "input:logo"
    },
    {
      "id": "web",
      "type": "convert",
      "input": "step:card",
      "output": "dist/card_web.jpg",
      "max_width": 720,
      "max_height": 720,
      "quality": 80,
      "strip_metadata": true
    }
  ]
}
```

## 预设尺寸

- `xiaohongshu`: `1080x1440`
- `wechat_cover`: `900x383`
- `square`: `1080x1080`
- `story_9x16`: `1080x1920`
- `product_square`: `1200x1200`
- `detail_long`: `1080x2160`
- `banner_16x9`: `1600x900`

预设集合：

- `creator-basic`：`xiaohongshu`、`wechat_cover`、`square`、`story_9x16`
- `ecommerce-basic`：`product_square`、`detail_long`、`banner_16x9`

## 真实烟测

如果本机已经配置好 `baoyu-image-gen`，可以执行端到端真实烟测：

```bash
make real-smoke
```

该流程会：

- 构建 `imgcli`
- 使用外部生成技能生成一组真实测试图
- 对这些图片执行 `inspect`
- 跑通 `compose`、`convert`、`variants`、`run`、`smartpad`、`topdf`、`stitch`
- 校验输出文件类型，并打印产物目录

常用覆盖项：

- `SKILL_DIR=/custom/skill/path make real-smoke`
- `SMOKE_ROOT=/tmp/custom-smokes make real-smoke`
- `RUN_ID=manual-check make real-smoke`

## 面向 Agent 的技能

- 技能文件：`skills/imgcli/SKILL.md`
- 仓库内还包含 `AGENTS.md`，用于约束工程规范和交付要求

## V2 状态

下一阶段能力规划见 [docs/v2-spec.md](/root/go/src/imgcli/docs/v2-spec.md)。
`compose` 的边界说明见 [docs/compose-boundary.md](/root/go/src/imgcli/docs/compose-boundary.md)。

V2 核心命令已经全部落地：`inspect`、`compose`、`convert`、`variants`、`run`。
`compose` 的定位仍然是“固定模板卡片渲染器”，不是命令行 Figma。

## 开发与质量门禁

必须通过：

```bash
gofmt -l .
go vet ./...
golangci-lint run
CGO_ENABLED=1 go test -count=1 ./...
make release-check
make build
```

常用命令：

```bash
make fmt
make vet
make lint
make test
make release-check
make build
make real-smoke
```
