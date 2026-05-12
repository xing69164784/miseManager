# Mise Manager

一个基于 Wails 3 构建的 **Mise** 可视化管理工具，提供直观的图形界面来管理你的开发工具版本。

## ✨ 功能特性

- 📦 **工具管理** - 查看已安装的工具及其版本
- 🚀 **快速安装** - 从 Mise 注册表搜索并安装新工具
- 🔄 **版本切换** - 轻松设置默认版本
- 🗑️ **卸载管理** - 安全卸载不需要的工具版本
- 📊 **诊断功能** - 运行 `mise doctor` 检查配置
- 💻 **命令终端** - 内置终端执行任意 Mise 命令
- 📝 **日志记录** - 自动记录所有操作日志

## 📋 系统要求

- **操作系统**: Windows / macOS / Linux
- **依赖**: 
  - [Mise](https://mise.jdx.dev/) (已安装并配置好)
  - Go 1.22+ (开发环境)
  - Node.js 18+ (开发环境)

## 🛠️ 安装方法

### 预编译版本

从 [Releases](https://github.com/xing69164784/miseManager/releases) 页面下载对应平台的二进制文件。

### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/xing69164784/miseManager.git
cd miseManager

# 安装依赖
npm install

# 切换到项目根目录
cd ..

# 开发模式运行
wails3 dev

# 构建生产版本
wails3 build
```

## 🎮 使用说明

### 工具管理

1. 打开应用后，左侧显示已安装的工具列表
2. 点击工具名称查看详情
3. 在详情面板中可以：
   - 安装新版本
   - 设置默认版本
   - 卸载指定版本

### 安装新工具

1. 点击右上角 ➕ 按钮
2. 在搜索框中输入工具名称
3. 从搜索结果中选择工具并点击安装

### 命令终端

1. 点击右上角终端图标
2. 输入 Mise 子命令（如 `ls`, `doctor`, `install go@1.21.0`）
3. 按 Enter 执行命令

### 日志页面

1. 点击下方「日志」标签
2. 查看所有操作记录
3. 支持手动刷新和清空日志

## 🛡️ 安全说明

- 所有命令在后台执行，不会弹出命令行窗口
- 日志文件存储在 `~/.mise-gui.log`
- 仅执行 Mise 相关命令，不会执行系统级操作

## 📁 项目结构

```
mise-manager/
├── frontend/           # Vue 3 前端代码
│   ├── src/
│   │   ├── components/ # UI 组件
│   │   ├── App.vue     # 主应用组件
│   │   └── main.js     # 入口文件
│   ├── bindings/       # Wails 自动生成的绑定
│   └── dist/           # 构建产物
├── main.go             # Go 后端入口
├── miseservice.go      # Mise 服务实现
├── go.mod              # Go 依赖配置
└── go.sum              # Go 依赖校验
```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License

## 🙏 致谢

- [Wails](https://wails.io/) - 跨平台桌面应用框架
- [Mise](https://mise.jdx.dev/) - 强大的版本管理工具
- [Vue.js](https://vuejs.org/) - 渐进式 JavaScript 框架