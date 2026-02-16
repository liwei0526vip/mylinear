# 后端脚手架配置扩展规格

## MODIFIED Requirements

### Requirement: 应用配置支持 JWT 相关配置项

系统 SHALL 在配置结构体中支持 JWT 密钥和过期时间配置。

#### Scenario: 加载默认 JWT 配置

- **WHEN** 未设置 JWT 相关环境变量
- **THEN** 配置加载使用默认值：
  - JWT_SECRET: 随机生成的 32 字节字符串（仅开发环境）
  - JWT_ACCESS_EXPIRY: 15 分钟
  - JWT_REFRESH_EXPIRY: 7 天

#### Scenario: 从环境变量加载 JWT 配置

- **GIVEN** 环境变量设置为：
  - JWT_SECRET: "my-secret-key"
  - JWT_ACCESS_EXPIRY: "30m"
  - JWT_REFRESH_EXPIRY: "24h"
- **WHEN** 配置加载
- **THEN** 配置结构体包含：
  - JWTSecret: "my-secret-key"
  - JWTAccessExpiry: 30 * time.Minute
  - JWTRefreshExpiry: 24 * time.Hour

#### Scenario: 生产环境必须设置 JWT_SECRET

- **GIVEN** 运行环境为生产（GIN_MODE=release）
- **WHEN** 未设置 JWT_SECRET 环境变量
- **THEN** 应用启动失败
- **AND** 日志输出 "生产环境必须设置 JWT_SECRET 环境变量"

### Requirement: 配置结构体新增字段

系统 SHALL 在 Config 结构体中新增以下字段：

```go
type Config struct {
    // 现有字段...

    // JWT 配置
    JWTSecret        string
    JWTAccessExpiry  time.Duration
    JWTRefreshExpiry time.Duration
}
```

#### Scenario: 配置结构体验证

- **WHEN** 配置加载完成
- **THEN** Config.JWTSecret 不为空（生产环境）或使用默认值（开发环境）
- **AND** Config.JWTAccessExpiry > 0
- **AND** Config.JWTRefreshExpiry > Config.JWTAccessExpiry
