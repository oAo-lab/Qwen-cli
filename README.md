# Qwen-cli

通义千问命令行客户端，支持多模型对话和角色切换。

## 功能特性

- 🤖 支持通义千问多模型对话
- 🎭 内置多种角色提示词
- 🌐 支持联网搜索
- 💾 支持对话历史保存
- ⚙️ 跨平台配置管理
- 🔧 环境变量配置支持

## 安装

### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/oAo-lab/Qwen-cli.git
cd Qwen-cli

# 构建项目
make build

# 安装到系统路径（可选）
make install
```

### 多平台构建

```bash
# 构建所有平台版本
make build-all

# 构建产物将输出到 dist/ 目录
```

## 快速开始

### 1. 初始化配置

```bash
ask init
```

这将在用户配置目录创建配置文件：
- **Windows**: `%USERPROFILE%\.config\ask\config.json`
- **macOS/Linux**: `~/.config/ask/config.json`

### 2. 配置 API 密钥

编辑配置文件设置您的 API 密钥：

```json
{
  "api_url": "https://dashscope.aliyuncs.com/compatible-mode/v1",
  "api_key": "your-api-key-here",
  "models": {
    "default": {
      "name": "qwen-turbo"
    }
  },
  "roles": {
    "default": "你是一个有用的AI助手，能够协助用户解决各种问题。"
  }
}
```

### 3. 开始对话

```bash
ask chat
```

## 使用方法

### 基本命令

```bash
# 初始化配置
ask init

# 开始聊天
ask chat

# AI命令助手
ask cmd

# 测试连接
ask test

# 调试模式
ask debug
```

### 聊天命令

在聊天模式下，支持以下命令：

- `/model` - 切换模型
- `/prompt` - 切换角色提示词
- `/online` - 开启/关闭联网搜索
- `/save` - 保存最后一次回复
- `/save -all` - 保存完整对话
- `exit` - 退出聊天

### AI命令助手

AI命令助手可以帮助您生成和执行系统命令：

```bash
# 启动命令助手（交互模式）
ask cmd

# 直接描述需求（非交互模式）
ask cmd "查看当前目录的文件"
```

使用流程：
1. 使用 `/cmd 命令描述` 来请求生成命令，或直接输入文本进行普通聊天
2. AI实时生成相应的系统命令（流式输出）
3. 确认是否执行该命令
4. 系统执行命令并显示结果
5. 命令执行结果会自动添加到AI上下文中，支持连续对话

示例需求：
- /cmd 查看当前目录的文件
- /cmd 创建一个名为test的目录
- /cmd 查看系统信息
- /cmd 查看端口8080是否被占用
- /cmd 查看磁盘使用情况

特色功能：
- 🔄 流式输出：AI生成命令时实时显示
- 🧠 上下文记忆：命令执行结果会被记住，支持基于结果的后续操作
- 🔄 连续对话：可以根据上一个命令的结果进行下一步操作
- 💬 智能区分：使用 `/cmd` 前缀明确区分命令请求和普通聊天
- 🖥️ 环境感知：自动检测操作系统、终端类型和当前环境，生成针对性的命令
- 📢 项目推广：当询问项目相关信息时，自动提供项目地址和介绍

### 其他命令

```bash
# 测试API连接
ask test

# 切换调试模式
ask debug
```

### 环境变量配置

您可以通过环境变量覆盖配置文件设置：

```bash
export ASK_API_URL="https://your-api-endpoint.com/v1"
export ASK_API_KEY="your-api-key"

ask chat
```

## 配置说明

### 配置文件结构

```json
{
  "api_url": "API 服务器地址",
  "api_key": "API 密钥",
  "models": {
    "default": {
      "name": "默认模型名称"
    },
    "custom-model": {
      "name": "自定义模型名称"
    }
  },
  "roles": {
    "default": "默认角色提示词",
    "programmer": "程序员角色提示词",
    "teacher": "老师角色提示词"
  }
}
```

### 支持的模型

- `qwen-turbo` - 快速响应模型
- `qwen-plus` - 平衡性能模型
- `qwen-max` - 高性能模型

### 内置角色

- `default` - 通用助手
- `programmer` - 程序员
- `translator` - 翻译
- `teacher` - 老师

## 开发

### 构建和测试

```bash
# 构建项目
make build

# 运行测试
make test

# 清理生成文件
make clean
```

### 发布流程

```bash
# 自动提交、创建版本并推送
make push

# 手动创建版本标签
make bump-version
```

## 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

## 贡献

欢迎提交 Issue 和 Pull Request！

## 链接

- [GitHub 仓库](https://github.com/oAo-lab/Qwen-cli)
- [发布页面](https://github.com/oAo-lab/Qwen-cli/releases)
