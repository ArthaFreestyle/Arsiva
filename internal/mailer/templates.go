package mailer

import (
	"bytes"
	"html/template"
	"strings"
)

// OTPEmail holds the copy that varies between OTP purposes (email verification
// vs. password reset). The visual layout is identical; only the wording changes.
type OTPEmail struct {
	Eyebrow      string // top-right label, e.g. "Keamanan Akun"
	Heading      string // main heading, e.g. "Verifikasi email kamu"
	Intro        string // one-line intro under the heading
	CodeLabel    string // label above the code, e.g. "Kode verifikasi kamu"
	Code         string // the 6-digit OTP
	ExpiryMins   int    // minutes until the code expires
	SecurityNote string // reassurance line for "didn't request this"
	Preheader    string // hidden inbox preview text
}

// otpHTMLTemplate is a table-based, inline-styled responsive email adapted from
// the "Arsiva OTP Email" design. It deliberately avoids JavaScript, external
// CSS, and web fonts so it renders consistently across email clients (Gmail,
// Outlook, Apple Mail). The interactive "copy code" button from the design is
// intentionally dropped — email clients do not execute scripts.
var otpHTMLTemplate = template.Must(template.New("otp").Parse(`<!DOCTYPE html>
<html lang="id">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="color-scheme" content="light dark">
<meta name="supported-color-schemes" content="light dark">
<title>{{.Heading}}</title>
</head>
<body style="margin:0; padding:0; width:100%; background-color:#eceae4;">
  <span style="display:none !important; visibility:hidden; opacity:0; color:transparent; height:0; width:0; overflow:hidden; mso-hide:all; font-size:1px; line-height:1px;">{{.Preheader}}&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;</span>

  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" style="background-color:#eceae4;">
    <tr>
      <td align="center" style="padding:40px 16px;">

        <table role="presentation" width="600" cellpadding="0" cellspacing="0" border="0" style="width:600px; max-width:600px;">

          <!-- Brand -->
          <tr>
            <td style="padding:4px 8px 24px 8px;">
              <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">
                <tr>
                  <td align="left" style="font-family:Georgia,'Times New Roman',serif; font-size:22px; font-weight:400; letter-spacing:0.02em; color:#2c2a26;">Arsiva</td>
                  <td align="right" style="font-family:Arial,Helvetica,sans-serif; font-size:11px; font-weight:400; letter-spacing:0.14em; text-transform:uppercase; color:#8a867c;">{{.Eyebrow}}</td>
                </tr>
              </table>
            </td>
          </tr>

          <!-- Card -->
          <tr>
            <td style="background-color:#ffffff; border:1px solid #e2ded4; border-radius:14px; padding:0;">
              <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">

                <tr>
                  <td style="padding:48px 48px 0 48px; font-family:Georgia,'Times New Roman',serif; font-size:26px; line-height:1.25; font-weight:400; color:#221f1b;">{{.Heading}}</td>
                </tr>

                <tr>
                  <td style="padding:16px 48px 0 48px; font-family:Arial,Helvetica,sans-serif; font-size:15px; line-height:1.6; color:#5c584f;">{{.Intro}}</td>
                </tr>

                <!-- Code block -->
                <tr>
                  <td style="padding:32px 48px 0 48px;">
                    <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" style="background-color:#f6f4ee; border:1px solid #e6e2d8; border-radius:10px;">
                      <tr>
                        <td align="center" style="padding:28px 24px 12px 24px; font-family:Arial,Helvetica,sans-serif; font-size:11px; font-weight:700; letter-spacing:0.16em; text-transform:uppercase; color:#918c81;">{{.CodeLabel}}</td>
                      </tr>
                      <tr>
                        <td align="center" style="padding:0 24px 28px 24px; font-family:'Courier New',Courier,monospace; font-size:40px; font-weight:700; letter-spacing:0.28em; color:#221f1b; mso-line-height-rule:exactly; line-height:44px;">{{.Code}}</td>
                      </tr>
                    </table>
                  </td>
                </tr>

                <tr>
                  <td style="padding:20px 48px 0 48px; font-family:Arial,Helvetica,sans-serif; font-size:14px; line-height:1.6; color:#5c584f;">Kode ini berlaku selama <strong style="color:#221f1b; font-weight:700;">{{.ExpiryMins}} menit</strong>. Demi keamanan, jangan bagikan kode ini kepada siapa pun &mdash; staf Arsiva tidak akan pernah memintanya.</td>
                </tr>

                <!-- Divider -->
                <tr>
                  <td style="padding:36px 48px 0 48px;">
                    <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0"><tr><td style="border-top:1px solid #eceae3; font-size:0; line-height:0;">&nbsp;</td></tr></table>
                  </td>
                </tr>

                <tr>
                  <td style="padding:24px 48px 48px 48px; font-family:Arial,Helvetica,sans-serif; font-size:13px; line-height:1.65; color:#8a867c;">{{.SecurityNote}}</td>
                </tr>

              </table>
            </td>
          </tr>

          <!-- Footer -->
          <tr>
            <td style="padding:28px 24px 8px 24px;" align="center">
              <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">
                <tr>
                  <td align="center" style="font-family:Arial,Helvetica,sans-serif; font-size:12px; line-height:1.6; color:#9c988e;">Butuh bantuan? Hubungi <a href="mailto:admin@arsiva.id" style="color:#8a5a2b; text-decoration:underline;">admin@arsiva.id</a></td>
                </tr>
                <tr>
                  <td align="center" style="padding-top:10px; font-family:Arial,Helvetica,sans-serif; font-size:11px; line-height:1.6; color:#b4b0a6;">Ini adalah pesan keamanan otomatis dari Arsiva.<br>&copy; 2026 Arsiva. Semua hak dilindungi.</td>
                </tr>
              </table>
            </td>
          </tr>

        </table>

      </td>
    </tr>
  </table>
</body>
</html>`))

// ResetLinkEmail holds the copy + link for the password-reset email. Unlike the
// OTP email it presents a clickable button (the user never types anything) plus
// a plain-URL fallback for clients that strip the button.
type ResetLinkEmail struct {
	Heading      string // main heading, e.g. "Atur ulang password kamu"
	Intro        string // one-line intro under the heading
	ButtonLabel  string // call-to-action text, e.g. "Reset Password"
	ResetURL     string // the full reset link (with token) the button points to
	ExpiryMins   int    // minutes until the link expires
	SecurityNote string // reassurance line for "didn't request this"
	Preheader    string // hidden inbox preview text
}

// resetLinkHTMLTemplate mirrors the OTP template's visual language (same palette,
// table-based layout, no JS/web fonts) but swaps the code block for a button.
var resetLinkHTMLTemplate = template.Must(template.New("reset").Parse(`<!DOCTYPE html>
<html lang="id">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="color-scheme" content="light dark">
<meta name="supported-color-schemes" content="light dark">
<title>{{.Heading}}</title>
</head>
<body style="margin:0; padding:0; width:100%; background-color:#eceae4;">
  <span style="display:none !important; visibility:hidden; opacity:0; color:transparent; height:0; width:0; overflow:hidden; mso-hide:all; font-size:1px; line-height:1px;">{{.Preheader}}&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;</span>

  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" style="background-color:#eceae4;">
    <tr>
      <td align="center" style="padding:40px 16px;">

        <table role="presentation" width="600" cellpadding="0" cellspacing="0" border="0" style="width:600px; max-width:600px;">

          <!-- Brand -->
          <tr>
            <td style="padding:4px 8px 24px 8px;">
              <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">
                <tr>
                  <td align="left" style="font-family:Georgia,'Times New Roman',serif; font-size:22px; font-weight:400; letter-spacing:0.02em; color:#2c2a26;">Arsiva</td>
                  <td align="right" style="font-family:Arial,Helvetica,sans-serif; font-size:11px; font-weight:400; letter-spacing:0.14em; text-transform:uppercase; color:#8a867c;">Keamanan Akun</td>
                </tr>
              </table>
            </td>
          </tr>

          <!-- Card -->
          <tr>
            <td style="background-color:#ffffff; border:1px solid #e2ded4; border-radius:14px; padding:0;">
              <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">

                <tr>
                  <td style="padding:48px 48px 0 48px; font-family:Georgia,'Times New Roman',serif; font-size:26px; line-height:1.25; font-weight:400; color:#221f1b;">{{.Heading}}</td>
                </tr>

                <tr>
                  <td style="padding:16px 48px 0 48px; font-family:Arial,Helvetica,sans-serif; font-size:15px; line-height:1.6; color:#5c584f;">{{.Intro}}</td>
                </tr>

                <!-- Button -->
                <tr>
                  <td style="padding:32px 48px 0 48px;">
                    <table role="presentation" cellpadding="0" cellspacing="0" border="0">
                      <tr>
                        <td align="center" bgcolor="#8a5a2b" style="border-radius:8px;">
                          <a href="{{.ResetURL}}" target="_blank" style="display:inline-block; padding:14px 32px; font-family:Arial,Helvetica,sans-serif; font-size:15px; font-weight:700; letter-spacing:0.02em; color:#ffffff; text-decoration:none; border-radius:8px;">{{.ButtonLabel}}</a>
                        </td>
                      </tr>
                    </table>
                  </td>
                </tr>

                <tr>
                  <td style="padding:24px 48px 0 48px; font-family:Arial,Helvetica,sans-serif; font-size:13px; line-height:1.6; color:#8a867c;">Kalau tombol di atas tidak berfungsi, salin dan tempel tautan berikut ke browser kamu:</td>
                </tr>
                <tr>
                  <td style="padding:8px 48px 0 48px; font-family:'Courier New',Courier,monospace; font-size:13px; line-height:1.5; color:#8a5a2b; word-break:break-all;"><a href="{{.ResetURL}}" target="_blank" style="color:#8a5a2b; text-decoration:underline; word-break:break-all;">{{.ResetURL}}</a></td>
                </tr>

                <tr>
                  <td style="padding:20px 48px 0 48px; font-family:Arial,Helvetica,sans-serif; font-size:14px; line-height:1.6; color:#5c584f;">Tautan ini berlaku selama <strong style="color:#221f1b; font-weight:700;">{{.ExpiryMins}} menit</strong> dan hanya bisa digunakan satu kali.</td>
                </tr>

                <!-- Divider -->
                <tr>
                  <td style="padding:36px 48px 0 48px;">
                    <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0"><tr><td style="border-top:1px solid #eceae3; font-size:0; line-height:0;">&nbsp;</td></tr></table>
                  </td>
                </tr>

                <tr>
                  <td style="padding:24px 48px 48px 48px; font-family:Arial,Helvetica,sans-serif; font-size:13px; line-height:1.65; color:#8a867c;">{{.SecurityNote}}</td>
                </tr>

              </table>
            </td>
          </tr>

          <!-- Footer -->
          <tr>
            <td style="padding:28px 24px 8px 24px;" align="center">
              <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">
                <tr>
                  <td align="center" style="font-family:Arial,Helvetica,sans-serif; font-size:12px; line-height:1.6; color:#9c988e;">Butuh bantuan? Hubungi <a href="mailto:admin@arsiva.id" style="color:#8a5a2b; text-decoration:underline;">admin@arsiva.id</a></td>
                </tr>
                <tr>
                  <td align="center" style="padding-top:10px; font-family:Arial,Helvetica,sans-serif; font-size:11px; line-height:1.6; color:#b4b0a6;">Ini adalah pesan keamanan otomatis dari Arsiva.<br>&copy; 2026 Arsiva. Semua hak dilindungi.</td>
                </tr>
              </table>
            </td>
          </tr>

        </table>

      </td>
    </tr>
  </table>
</body>
</html>`))

// RenderResetLinkHTML renders the password-reset email to an HTML string.
func RenderResetLinkHTML(data ResetLinkEmail) (string, error) {
	var buf bytes.Buffer
	if err := resetLinkHTMLTemplate.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderResetLinkText builds a plain-text fallback mirroring the HTML content.
func RenderResetLinkText(data ResetLinkEmail) string {
	var b strings.Builder
	b.WriteString("Halo,\n\n")
	b.WriteString(data.Intro)
	b.WriteString("\n\n")
	b.WriteString(data.ButtonLabel)
	b.WriteString(":\n")
	b.WriteString(data.ResetURL)
	b.WriteString("\n\n")
	b.WriteString("Tautan ini berlaku selama ")
	b.WriteString(itoa(data.ExpiryMins))
	b.WriteString(" menit dan hanya bisa digunakan satu kali.\n\n")
	b.WriteString(data.SecurityNote)
	b.WriteString("\n\nButuh bantuan? Hubungi admin@arsiva.id\n\nSalam,\nTim Arsiva")
	return b.String()
}

// GroupInviteEmail holds the copy + link for a group-invitation email. It shares
// the reset email's button-based visual language.
type GroupInviteEmail struct {
	GroupName      string // the group the student is invited to
	InviterName    string // the guru who sent the invite
	PersonalNote   string // optional free-text message from the guru ("" hides it)
	ButtonLabel    string // call-to-action text, e.g. "Gabung Grup"
	InviteURL      string // full join link (with token)
	ExpiryDays     int    // days until the invite link expires
	SecurityNote   string // reassurance line for "didn't expect this"
	Preheader      string // hidden inbox preview text
}

// groupInviteHTMLTemplate mirrors the reset-link template (same palette, button)
// but frames the copy as a group invitation and optionally shows a personal note.
var groupInviteHTMLTemplate = template.Must(template.New("invite").Parse(`<!DOCTYPE html>
<html lang="id">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="color-scheme" content="light dark">
<meta name="supported-color-schemes" content="light dark">
<title>Undangan Grup {{.GroupName}}</title>
</head>
<body style="margin:0; padding:0; width:100%; background-color:#eceae4;">
  <span style="display:none !important; visibility:hidden; opacity:0; color:transparent; height:0; width:0; overflow:hidden; mso-hide:all; font-size:1px; line-height:1px;">{{.Preheader}}&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;&#8199;</span>

  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" style="background-color:#eceae4;">
    <tr>
      <td align="center" style="padding:40px 16px;">

        <table role="presentation" width="600" cellpadding="0" cellspacing="0" border="0" style="width:600px; max-width:600px;">

          <!-- Brand -->
          <tr>
            <td style="padding:4px 8px 24px 8px;">
              <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">
                <tr>
                  <td align="left" style="font-family:Georgia,'Times New Roman',serif; font-size:22px; font-weight:400; letter-spacing:0.02em; color:#2c2a26;">Arsiva</td>
                  <td align="right" style="font-family:Arial,Helvetica,sans-serif; font-size:11px; font-weight:400; letter-spacing:0.14em; text-transform:uppercase; color:#8a867c;">Undangan Grup</td>
                </tr>
              </table>
            </td>
          </tr>

          <!-- Card -->
          <tr>
            <td style="background-color:#ffffff; border:1px solid #e2ded4; border-radius:14px; padding:0;">
              <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">

                <tr>
                  <td style="padding:48px 48px 0 48px; font-family:Georgia,'Times New Roman',serif; font-size:26px; line-height:1.25; font-weight:400; color:#221f1b;">Kamu diundang ke grup {{.GroupName}}</td>
                </tr>

                <tr>
                  <td style="padding:16px 48px 0 48px; font-family:Arial,Helvetica,sans-serif; font-size:15px; line-height:1.6; color:#5c584f;"><strong style="color:#221f1b;">{{.InviterName}}</strong> mengundang kamu untuk bergabung ke grup belajar <strong style="color:#221f1b;">{{.GroupName}}</strong> di Arsiva.</td>
                </tr>
{{if .PersonalNote}}
                <!-- Personal note -->
                <tr>
                  <td style="padding:24px 48px 0 48px;">
                    <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" style="background-color:#f6f4ee; border:1px solid #e6e2d8; border-radius:10px;">
                      <tr>
                        <td style="padding:18px 22px; font-family:Arial,Helvetica,sans-serif; font-size:14px; line-height:1.65; color:#5c584f; font-style:italic;">&ldquo;{{.PersonalNote}}&rdquo;</td>
                      </tr>
                    </table>
                  </td>
                </tr>
{{end}}
                <!-- Button -->
                <tr>
                  <td style="padding:32px 48px 0 48px;">
                    <table role="presentation" cellpadding="0" cellspacing="0" border="0">
                      <tr>
                        <td align="center" bgcolor="#8a5a2b" style="border-radius:8px;">
                          <a href="{{.InviteURL}}" target="_blank" style="display:inline-block; padding:14px 32px; font-family:Arial,Helvetica,sans-serif; font-size:15px; font-weight:700; letter-spacing:0.02em; color:#ffffff; text-decoration:none; border-radius:8px;">{{.ButtonLabel}}</a>
                        </td>
                      </tr>
                    </table>
                  </td>
                </tr>

                <tr>
                  <td style="padding:24px 48px 0 48px; font-family:Arial,Helvetica,sans-serif; font-size:13px; line-height:1.6; color:#8a867c;">Kalau tombol di atas tidak berfungsi, salin dan tempel tautan berikut ke browser kamu:</td>
                </tr>
                <tr>
                  <td style="padding:8px 48px 0 48px; font-family:'Courier New',Courier,monospace; font-size:13px; line-height:1.5; color:#8a5a2b; word-break:break-all;"><a href="{{.InviteURL}}" target="_blank" style="color:#8a5a2b; text-decoration:underline; word-break:break-all;">{{.InviteURL}}</a></td>
                </tr>

                <tr>
                  <td style="padding:20px 48px 0 48px; font-family:Arial,Helvetica,sans-serif; font-size:14px; line-height:1.6; color:#5c584f;">Undangan ini berlaku selama <strong style="color:#221f1b; font-weight:700;">{{.ExpiryDays}} hari</strong>. Kamu perlu masuk atau mendaftar sebagai siswa untuk bergabung.</td>
                </tr>

                <!-- Divider -->
                <tr>
                  <td style="padding:36px 48px 0 48px;">
                    <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0"><tr><td style="border-top:1px solid #eceae3; font-size:0; line-height:0;">&nbsp;</td></tr></table>
                  </td>
                </tr>

                <tr>
                  <td style="padding:24px 48px 48px 48px; font-family:Arial,Helvetica,sans-serif; font-size:13px; line-height:1.65; color:#8a867c;">{{.SecurityNote}}</td>
                </tr>

              </table>
            </td>
          </tr>

          <!-- Footer -->
          <tr>
            <td style="padding:28px 24px 8px 24px;" align="center">
              <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">
                <tr>
                  <td align="center" style="font-family:Arial,Helvetica,sans-serif; font-size:12px; line-height:1.6; color:#9c988e;">Butuh bantuan? Hubungi <a href="mailto:admin@arsiva.id" style="color:#8a5a2b; text-decoration:underline;">admin@arsiva.id</a></td>
                </tr>
                <tr>
                  <td align="center" style="padding-top:10px; font-family:Arial,Helvetica,sans-serif; font-size:11px; line-height:1.6; color:#b4b0a6;">Ini adalah pesan otomatis dari Arsiva.<br>&copy; 2026 Arsiva. Semua hak dilindungi.</td>
                </tr>
              </table>
            </td>
          </tr>

        </table>

      </td>
    </tr>
  </table>
</body>
</html>`))

// RenderGroupInviteHTML renders the group-invitation email to an HTML string.
func RenderGroupInviteHTML(data GroupInviteEmail) (string, error) {
	var buf bytes.Buffer
	if err := groupInviteHTMLTemplate.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderGroupInviteText builds a plain-text fallback mirroring the HTML content.
func RenderGroupInviteText(data GroupInviteEmail) string {
	var b strings.Builder
	b.WriteString("Halo,\n\n")
	b.WriteString(data.InviterName)
	b.WriteString(" mengundang kamu untuk bergabung ke grup belajar \"")
	b.WriteString(data.GroupName)
	b.WriteString("\" di Arsiva.\n\n")
	if data.PersonalNote != "" {
		b.WriteString("Pesan dari pengundang:\n\"")
		b.WriteString(data.PersonalNote)
		b.WriteString("\"\n\n")
	}
	b.WriteString(data.ButtonLabel)
	b.WriteString(":\n")
	b.WriteString(data.InviteURL)
	b.WriteString("\n\n")
	b.WriteString("Undangan ini berlaku selama ")
	b.WriteString(itoa(data.ExpiryDays))
	b.WriteString(" hari. Kamu perlu masuk atau mendaftar sebagai siswa untuk bergabung.\n\n")
	b.WriteString(data.SecurityNote)
	b.WriteString("\n\nButuh bantuan? Hubungi admin@arsiva.id\n\nSalam,\nTim Arsiva")
	return b.String()
}

// RenderOTPHTML renders the OTP email to an HTML string.
func RenderOTPHTML(data OTPEmail) (string, error) {
	var buf bytes.Buffer
	if err := otpHTMLTemplate.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderOTPText builds a plain-text fallback that mirrors the HTML content, for
// clients that do not render HTML and for spam-filter friendliness.
func RenderOTPText(data OTPEmail) string {
	var b strings.Builder
	b.WriteString("Halo,\n\n")
	b.WriteString(data.Intro)
	b.WriteString("\n\n")
	b.WriteString(data.CodeLabel)
	b.WriteString(": ")
	b.WriteString(data.Code)
	b.WriteString("\n\n")
	b.WriteString("Kode ini berlaku selama ")
	b.WriteString(itoa(data.ExpiryMins))
	b.WriteString(" menit. Demi keamanan, jangan bagikan kode ini kepada siapa pun — staf Arsiva tidak akan pernah memintanya.\n\n")
	b.WriteString(data.SecurityNote)
	b.WriteString("\n\nButuh bantuan? Hubungi admin@arsiva.id\n\nSalam,\nTim Arsiva")
	return b.String()
}

// itoa avoids pulling strconv into this file for a single conversion.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
