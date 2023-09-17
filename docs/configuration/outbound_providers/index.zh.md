# 出站提供者

### 结构

```json
{
  "outbounds": [
    {
      "type": "",
      "tag": "",
      "path": "",
      "healthcheck_url": "https://www.gstatic.com/generate_204",
    }
  ]
}
```

### 字段

| 类型   | 格式            |
|--------|----------------|
| `http` | [HTTP](./http) |
| `file` | [File](./file) |

#### tag

出站提供者的标签。

#### path

==必填==

出站提供者本地文件路径。

#### healthcheck_url

出站提供者健康检查的地址。

默认为 `https://www.gstatic.com/generate_204`。
