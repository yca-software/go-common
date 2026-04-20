# Email

Email sending via **Resend** and HTML template rendering from the filesystem. Used by go-api for auth (verification, password reset), invitations, and other transactional emails.

## Structure

| File        | Role                                                                 |
| ----------- | -------------------------------------------------------------------- |
| **main.go** | `Config`, `EmailService` interface, `NewEmailService(cfg)`, `SendEmail`, `PrepareEmailBody`. |
| **mock.go** | Mock of `EmailService` for tests.                                    |
| **main_test.go** | Tests for template loading and body preparation (no live send). |

## Config

- **ResendAPIKey** — Resend API key.
- **FromEmail**, **FromName** — Sender identity.
- **TemplatesPath** — Directory containing `.html` templates (e.g. `welcome.html`). Template names must not contain `..` or path separators.

## Usage

```go
import "github.com/yca-software/go-common/email"

svc := email.NewEmailService(&email.Config{
    ResendAPIKey:  cfg.ResendAPIKey,
    FromEmail:     cfg.FromEmail,
    FromName:      cfg.FromName,
    TemplatesPath: "/path/to/templates",
})
body, _ := svc.PrepareEmailBody("welcome", data)
_ = svc.SendEmail("user@example.com", "Welcome", body)
```

Inject `EmailService` in services that send mail; use the mock in tests.
