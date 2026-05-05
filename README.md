# SmartPing

[English](README_EN.md) | 中文

SmartPing 是一个开源的网络延迟监控工具，支持多地区多运营商的 Ping 监测，通过可视化地图展示全国各地的网络延迟情况。

![SmartPing界面](screenshot.png)

## 功能特性

- 🗺️ **中国地图可视化** - 在地图上直观展示各省份延迟情况
- 📡 **多运营商支持** - 支持电信、联通、移动三大运营商
- 📊 **历史数据** - 查看历史延迟曲线，分析网络质量
- 🔄 **定时检测** - 每分钟自动检测，无需人工干预
- 📈 **排名展示** - 按延迟高低展示省份排名
- ⚙️ **配置灵活** - 支持自定义监控节点

## 快速开始

### 环境要求

- Go 1.16+
- Git

### 编译

```bash
# 克隆项目
git clone https://github.com/smartping/smartping.git
cd smartping

# 编译
cd src
go build -o smartping ..
```

### 配置

复制配置文件模板并编辑：

```bash
cp conf/config.json.example conf/config.json
vi conf/config.json
```

配置文件中可以添加要监控的 IP 地址和端口。

### 运行

```bash
# Linux/Mac
./smartping

# Windows
smartping.exe
```

访问 http://localhost:8899

## 项目结构

```
smartping/
├── html/              # 前端页面
│   ├── index.html    # 中国地图监控页面
│   └── assets/       # 静态资源 (CSS, JS, 地图数据)
├── conf/             # 配置文件
│   └── config.json.example  # 配置模板
├── src/              # Go 源代码
│   ├── smartping.go  # 主程序入口
│   ├── http/         # HTTP 服务器
│   ├── funcs/        # 核心功能函数
│   └── g/            # 全局变量和配置
├── db/               # SQLite 数据库
├── database/         # 数据库初始化脚本
└── logs/            # 日志目录
```

## 配置说明

`config.json` 主要配置项：

| 配置项 | 说明 |
|--------|------|
| `Port` | HTTP 服务端口，默认 8899 |
| `Base.Archive` | 数据归档天数 |
| `Base.Refresh` | 刷新间隔（秒） |
| `Base.Timeout` | Ping 超时时间（秒） |
| `Chinamap` | 中国地图监控配置，按省份和运营商定义监控节点 |

### Chinamap 配置示例

```json
"Chinamap": {
    "北京": {
        "ctcc": {
            "北京": ["220.181.121.33"]
        },
        "cucc": {
            "北京": ["61.49.140.201"]
        },
        "cmcc": {
            "北京": ["117.136.0.126"]
        }
    }
}
```

### 添加监控节点（IP）

在 `Chinamap` 中添加监控节点，格式为：

```json
"省份": {
    "运营商代码": {
        "城市": ["IP1", "IP2"]
    }
}
```

**运营商代码：**
- `ctcc` - 中国电信
- `cucc` - 中国联通
- `cmcc` - 中国移动

**示例：添加四川电信和联通节点**

```json
"Chinamap": {
    "北京": {
        "ctcc": { "北京": ["220.181.121.33"] },
        "cucc": { "北京": ["61.49.140.201"] },
        "cmcc": { "北京": ["117.136.0.126"] }
    },
    "四川": {
        "ctcc": {
            "成都": ["182.140.142.199"],
            "绵阳": ["118.112.1.1"]
        },
        "cucc": {
            "成都": ["119.6.100.1"]
        },
        "cmcc": {
            "成都": ["183.223.15.58"],
            "雅安": ["117.176.244.143"]
        }
    }
}
```

**注意事项：**
- 同一城市可以添加多个 IP，会计算平均延迟
- 建议每个省份至少配置 3 个运营商的节点
- IP 地址需要是可以被 Ping 的目标设备

## 数据说明

- **延迟颜色**：
  - 绿色 (≤50ms) - 优秀
  - 浅绿 (≤100ms) - 良好
  - 黄色 (≤200ms) - 一般
  - 橙色 (≤250ms) - 较差
  - 红色 (>250ms 或超时) - 很差

## 常见问题

**Q: 地图不显示怎么办？**
A: 检查浏览器控制台是否有错误，确保 `html/assets/map/china.json` 文件存在。

**Q: 如何添加新的监控节点？**
A: 在 `config.json` 的 `Chinamap` 部分添加新的省份和 IP 配置。

**Q: 数据保存在哪里？**
A: 数据存储在 `db/` 目录下的 SQLite 数据库中。

## 贡献代码

欢迎提交 Issue 和 Pull Request！

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/xxx`)
3. 提交更改 (`git commit -m 'Add xxx'`)
4. 推送到分支 (`git push origin feature/xxx`)
5. 创建 Pull Request

## 致谢

本项目基于 [SmartPing](https://github.com/smartping/smartping) 修改而来。

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 致谢

- [ECharts](https://echarts.apache.org/) - 数据可视化库
- [Go](https://golang.org/) - Go 语言
