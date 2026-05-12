package main

import (
	"fmt"
	"keyclub-api/app"
	"net/smtp"
	"strings"
)

func main() {
	fmt.Println("Hello, World!")
	config := app.LoadConfig()

	from := config.SMTPUser
	to := []string{"nikolay.li2008@gmail.com"}
	// If your server needs auth:
	auth := smtp.PlainAuth("", from, config.SMTPPassword, config.SMTPHost)
	subject := "Hello (HTML)"
	htmlBody := `
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width,initial-scale=1" />
    <meta name="x-apple-disable-message-reformatting" />
    <title>Email</title>
  </head>
  <body style="margin:0;padding:0;background:#F7F8F4;">
    <!-- Preheader (hidden) -->
    <div style="display:none;max-height:0;overflow:hidden;opacity:0;color:transparent;">
      A short summary of this email goes here.
    </div>
    <table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0" style="background:#F7F8F4;">
      <tr>
        <td align="center" style="padding:28px 16px;">
          <!-- Card -->
          <table role="presentation" width="600" cellspacing="0" cellpadding="0" border="0" style="width:600px;max-width:600px;background:#FFFFFF;border:1px solid #E6E8E1;border-radius:16px;overflow:hidden;">
            <!-- Accent top bar (10%) -->
            <tr>
              <td style="background:#E6F76A;height:10px;line-height:10px;font-size:0;">&nbsp;</td>
            </tr>
            <!-- Header block (dark) -->
            <tr>
              <td style="background:#2B2F45;padding:28px 24px 18px 24px;text-align:center;">
                <div style="font-family:system-ui,-apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;font-size:22px;line-height:28px;font-weight:700;color:#FFFFFF;">
                  Centered Header
                </div>
                <div style="margin-top:8px;font-family:system-ui,-apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;font-size:14px;line-height:20px;font-weight:500;color:#E6F76A;">
                  Subheader goes here (short + clear)
                </div>
              </td>
            </tr>
            <!-- Content (light area inside card) -->
            <tr>
              <td style="padding:22px 24px 26px 24px;background:#FFFFFF;">
                <div style="font-family:system-ui,-apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;font-size:15px;line-height:22px;color:#111111;">
                  <p style="margin:0 0 12px 0;">
                    Hi Nikolay,
                  </p>
                  <p style="margin:0 0 12px 0;">
                    This is a simple Gmail-friendly HTML email layout using your palette. Keep this section for your main message.
                    Later, you can add a button under the paragraph.
                  </p>
                  <p style="margin:0;">
                    — Your App
                  </p>
                </div>
                <!-- Button placeholder (disabled for now)
                <div style="margin-top:18px;text-align:left;">
                  <a href="https://example.com"
                     style="display:inline-block;background:#2B2F45;color:#FFFFFF;text-decoration:none;
                            font-family:system-ui,-apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;
                            font-size:14px;line-height:16px;font-weight:700;padding:12px 16px;border-radius:12px;">
                    Call to action
                  </a>
                </div>
                -->
              </td>
            </tr>
            <!-- Footer -->
            <tr>
              <td style="padding:14px 24px 18px 24px;background:#FFFFFF;border-top:1px solid #EEF0EA;">
                <div style="font-family:system-ui,-apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;font-size:12px;line-height:18px;color:#666666;text-align:center;">
                  If you didn’t request this, you can ignore this email.
                </div>
              </td>
            </tr>
          </table>
          <!-- /Card -->
        </td>
      </tr>
    </table>
  </body>
</html>
	`
	// Build RFC 5322 message
	msg := strings.Join([]string{
		fmt.Sprintf("From: %s", from),
		fmt.Sprintf("To: %s", strings.Join(to, ", ")),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		`Content-Type: text/html; charset="UTF-8"`,
		"Content-Transfer-Encoding: 7bit",
		"", // blank line between headers and body
		htmlBody,
	}, "\r\n")
	addr := config.SMTPAddress
	if err := smtp.SendMail(addr, auth, from, to, []byte(msg)); err != nil {
		panic(err)
	}
}
