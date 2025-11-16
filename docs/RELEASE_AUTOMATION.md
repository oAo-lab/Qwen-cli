# Release 自动化说明

## 概述

本项目已配置了自动化的发布流程，当主分支有代码变更时会自动创建新版本并发布。

## 工作流程

### 1. 自动变更检测

- 每次推送到 `main` 或 `master` 分支时，会触发变更检测
- 系统会比较最新发布标签与当前提交之间的差异
- 如果检测到变更，会自动生成新的版本号（补丁版本递增）

### 2. 版本号生成

- 如果没有历史发布标签，初始版本为 `v0.1.0`
- 如果有历史发布标签，会自动递增补丁版本号（例如 `v1.2.3` → `v1.2.4`）

### 3. 自动发布流程

1. 检测到变更后，自动创建新的 Git 标签
2. 推送标签到远程仓库
3. 触发 GoReleaser 进行构建和发布
4. 生成多平台二进制文件（Linux、Windows、macOS）
5. 创建 GitHub Release 并上传构建产物

## 手动触发发布

### 方法一：推送标签

```bash
git tag v1.2.3
git push origin v1.2.3
```

### 方法二：使用测试工作流

1. 进入 GitHub Actions 页面
2. 选择 "Test Release Workflow" 工作流
3. 点击 "Run workflow"
4. 选择测试模式：
   - `dry-run`: 仅测试逻辑，不实际发布
   - `full-test`: 完整测试构建流程

## 配置文件说明

### `.github/workflows/release.yml`

主要发布工作流，包含三个作业：

1. `check-changes`: 检测变更并生成版本号
2. `create-release`: 创建标签并发布
3. `tag-release`: 兼容手动标签推送

### `.github/workflows/test-release.yml`

测试工作流，用于验证发布逻辑：

1. 测试变更检测逻辑
2. 测试版本号生成
3. 干运行或完整测试 GoReleaser

### `.goreleaser.yaml`

GoReleaser 配置文件，定义了：

- 构建目标平台（Linux、Windows、macOS）
- 架构（amd64、arm64）
- 发布配置
- 变更日志过滤规则

## 注意事项

1. **权限要求**: 确保 GitHub Actions 有 `contents: write` 权限
2. **分支保护**: 如果主分支有保护规则，可能需要调整权限
3. **版本策略**: 当前使用简单的补丁版本递增，可根据需要调整
4. **变更日志**: 排除了 `test:`、`ci:`、`docs:` 等前缀的提交

## 自定义配置

### 修改版本递增策略

在 `.github/workflows/release.yml` 的 `Generate next version` 步骤中修改版本号生成逻辑：

```bash
# 主版本递增
MAJOR=$((MAJOR + 1))
MINOR=0
PATCH=0

# 次版本递增
MINOR=$((MINOR + 1))
PATCH=0

# 补丁版本递增（当前）
PATCH=$((PATCH + 1))
```

### 修改触发分支

在 `on.push.branches` 中修改触发自动发布的分支：

```yaml
push:
  branches:
    - main
    - master
    - develop  # 添加 develop 分支
```

### 修改变更日志过滤

在 `.goreleaser.yaml` 的 `changelog.filters.exclude` 中调整排除的提交类型：

```yaml
changelog:
  filters:
    exclude:
      - '^test:'
      - '^ci:'
      # 移除 '^docs:' 以包含文档更新
```

## 故障排除

### 发布失败

1. 检查 GitHub Actions 日志
2. 确认仓库权限设置
3. 验证 `.goreleaser.yaml` 配置

### 版本号冲突

1. 检查是否已存在相同标签
2. 手动删除冲突标签后重试

### 构建失败

1. 检查 Go 版本兼容性
2. 验证依赖项是否正确
3. 查看构建日志中的错误信息