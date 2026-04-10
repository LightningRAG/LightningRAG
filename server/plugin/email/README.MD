## LightningRAG email plugin

**简体中文：** [README_zh.md](./README_zh.md)

Author: LightningRAG

### Setup

1. In `initialize/router.go`, at the end of `Routers`, register the plugin with `PluginInit(PrivateGroup, email.CreateEmailPlug(...))`.
2. Pass settings from `global.LRAG_CONFIG.Email` or hard-code SMTP fields (`To`, `From`, `Host`, `Secret`, `Nickname`, `Port`, `IsSSL`, `IsLoginAuth`).

### Config struct (global)

`From` / `Secret` are typically SMTP username and app password (not the mailbox login password if the provider issues an SMTP token).

```go
type Email struct {
    To          string // Recipients, comma-separated; often used for error alerts
    From        string // Sender address
    Host        string // e.g. smtp.qq.com
    Secret      string // SMTP auth secret
    Nickname    string // Optional display name
    Port        int    // Often 465 with SSL
    IsSSL       bool
    IsLoginAuth bool   // Some providers (IBM, Microsoft, etc.)
}
```

### Request body (send API)

```go
type Email struct {
    To      string `json:"to"`
    Subject string `json:"subject"`
    Body    string `json:"body"`
}
```

### Helpers (`utils`)

- `EmailTest(subject, body)` — send a test message  
- `ErrorToEmail(subject, body)` — error notification path  
- `Email(toCommaSeparated, subject, body)` — general send  

### HTTP routes (Swagger)

- `POST /email/emailTest`  
- `POST /email/emailSend` — body: `to`, `subject`, `body`  
