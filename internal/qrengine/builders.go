package qrengine

import (
	"fmt"
	"strings"
	"time"
)

func escapeWiFi(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `;`, `\;`, `:`, `\:`, `,`, `\,`, `"`, `\"`)
	return r.Replace(s)
}

func escapeVCard(s string) string {
	r := strings.NewReplacer("\\", "\\\\", ";", "\\;", ",", "\\,", "\n", "\\n")
	return r.Replace(s)
}

func escapeICal(s string) string {
	r := strings.NewReplacer("\\", "\\\\", ";", "\\;", ",", "\\,", "\n", "\\n")
	return r.Replace(s)
}

type BuildContentRequest struct {
	Type   ContentType       `json:"type"`
	Fields map[string]string `json:"fields"`
}

func missing(fields map[string]string, keys ...string) []string {
	var out []string
	for _, k := range keys {
		if strings.TrimSpace(fields[k]) == "" {
			out = append(out, k)
		}
	}
	return out
}

func BuildContent(req BuildContentRequest) (string, error) {
	f := req.Fields
	for k, v := range f {
		f[k] = strings.TrimSpace(v)
	}
	switch req.Type {
	case TypeText:
		if s := f["text"]; s != "" {
			return s, nil
		}
		return "", fmt.Errorf("missing required field: text")
	case TypeURL:
		s := f["url"]
		if s == "" {
			return "", fmt.Errorf("missing required field: url")
		}
		if !strings.Contains(s, "://") && !strings.HasPrefix(s, "mailto:") {
			s = "https://" + s
		}
		return s, nil
	case TypeWiFi:
		if m := missing(f, "ssid"); len(m) > 0 {
			return "", fmt.Errorf("missing required field: ssid")
		}
		enc := strings.ToUpper(f["enc"])
		pass := f["password"]
		switch enc {
		case "":
			if pass != "" {
				enc = "WPA"
			} else {
				enc = "nopass"
			}
		case "WEP", "WPA", "WPA2", "WPA/WPA2":
			if enc == "WPA2" || enc == "WPA/WPA2" {
				enc = "WPA"
			}
		case "NOPASS":
			enc = "nopass"
		default:
			return "", fmt.Errorf("invalid encryption: %s (use WPA, WEP or nopass)", enc)
		}
		var b strings.Builder
		b.WriteString("WIFI:T:" + enc + ";S:" + escapeWiFi(f["ssid"]) + ";")
		if pass != "" && enc != "nopass" {
			b.WriteString("P:" + escapeWiFi(pass) + ";")
		}
		if strings.EqualFold(f["hidden"], "true") || strings.EqualFold(f["hidden"], "yes") {
			b.WriteString("H:true;")
		}
		b.WriteString(";")
		return b.String(), nil
	case TypeEmail:
		if m := missing(f, "to"); len(m) > 0 {
			return "", fmt.Errorf("missing required field: to")
		}
		if strings.HasPrefix(strings.ToLower(f["to"]), "mailto:") {
			return f["to"], nil
		}
		s := "mailto:" + f["to"]
		var q []string
		if sub := f["subject"]; sub != "" {
			q = append(q, "subject="+urlQueryEscape(sub))
		}
		if body := f["body"]; body != "" {
			q = append(q, "body="+urlQueryEscape(body))
		}
		if len(q) > 0 {
			s += "?" + strings.Join(q, "&")
		}
		return s, nil
	case TypeSMS:
		if m := missing(f, "number"); len(m) > 0 {
			return "", fmt.Errorf("missing required field: number")
		}
		if msg := f["message"]; msg != "" {
			return "SMSTO:" + f["number"] + ":" + msg, nil
		}
		return "SMSTO:" + f["number"] + ":", nil
	case TypePhone:
		if m := missing(f, "number"); len(m) > 0 {
			return "", fmt.Errorf("missing required field: number")
		}
		n := f["number"]
		if strings.HasPrefix(strings.ToLower(n), "tel:") {
			return n, nil
		}
		return "tel:" + n, nil
	case TypeVCard:
		fn := f["fn"]
		first, last := f["first"], f["last"]
		if fn == "" {
			fn = strings.TrimSpace(first + " " + last)
		}
		if fn == "" {
			return "", fmt.Errorf("missing required field: fn (or first/last)")
		}
		var b strings.Builder
		b.WriteString("BEGIN:VCARD\r\nVERSION:3.0\r\n")
		b.WriteString("N:" + escapeVCard(last) + ";" + escapeVCard(first) + ";;;\r\n")
		b.WriteString("FN:" + escapeVCard(fn) + "\r\n")
		for _, pair := range [][2]string{
			{"org", "ORG"}, {"title", "TITLE"}, {"tel", "TEL;TYPE=CELL"},
			{"email", "EMAIL"}, {"url", "URL"},
		} {
			if v := f[pair[0]]; v != "" {
				b.WriteString(pair[1] + ":" + escapeVCard(v) + "\r\n")
			}
		}
		if adr := f["adr"]; adr != "" {
			b.WriteString("ADR;TYPE=WORK:;;" + escapeVCard(adr) + ";;;;\r\n")
		}
		if note := f["note"]; note != "" {
			b.WriteString("NOTE:" + escapeVCard(note) + "\r\n")
		}
		b.WriteString("END:VCARD")
		return b.String(), nil
	case TypeGeo:
		if m := missing(f, "lat", "lng"); len(m) > 0 {
			return "", fmt.Errorf("missing required fields: lat, lng")
		}
		if !isFloat(f["lat"]) || !isFloat(f["lng"]) {
			return "", fmt.Errorf("lat/lng must be numeric coordinates")
		}
		return fmt.Sprintf("geo:%s,%s", f["lat"], f["lng"]), nil
	case TypeEvent:
		if m := missing(f, "title", "start"); len(m) > 0 {
			return "", fmt.Errorf("missing required fields: title, start")
		}
		startDT, err := parseEventTime(f["start"])
		if err != nil {
			return "", fmt.Errorf("invalid start time: %w", err)
		}
		var endDT string
		if f["end"] != "" {
			endDT, err = parseEventTime(f["end"])
			if err != nil {
				return "", fmt.Errorf("invalid end time: %w", err)
			}
		}
		var b strings.Builder
		b.WriteString("BEGIN:VEVENT\r\n")
		b.WriteString("SUMMARY:" + escapeICal(f["title"]) + "\r\n")
		b.WriteString("DTSTART:" + startDT + "\r\n")
		if endDT != "" {
			b.WriteString("DTEND:" + endDT + "\r\n")
		}
		if loc := f["location"]; loc != "" {
			b.WriteString("LOCATION:" + escapeICal(loc) + "\r\n")
		}
		if desc := f["description"]; desc != "" {
			b.WriteString("DESCRIPTION:" + escapeICal(desc) + "\r\n")
		}
		b.WriteString("END:VEVENT")
		return b.String(), nil
	case TypeSocial:
		if m := missing(f, "platform", "handle"); len(m) > 0 {
			return "", fmt.Errorf("missing required fields: platform, handle")
		}
		platform := strings.ToLower(strings.TrimSpace(f["platform"]))
		handle := strings.TrimPrefix(strings.TrimSpace(f["handle"]), "@")
		if handle == "" {
			return "", fmt.Errorf("handle cannot be empty")
		}
		switch platform {
		case "x", "twitter":
			return "https://x.com/" + handle, nil
		case "instagram":
			return "https://instagram.com/" + handle, nil
		case "github":
			return "https://github.com/" + handle, nil
		case "linkedin":
			return "https://linkedin.com/in/" + handle, nil
		case "youtube":
			return "https://youtube.com/@" + handle, nil
		case "tiktok":
			return "https://tiktok.com/@" + handle, nil
		case "telegram":
			return "https://t.me/" + handle, nil
		case "whatsapp":
			phone := strings.Map(func(r rune) rune {
				if r >= '0' && r <= '9' {
					return r
				}
				return -1
			}, handle)
			if phone == "" {
				return "", fmt.Errorf("whatsapp requires a phone number, not a username")
			}
			return "https://wa.me/" + phone, nil
		default:
			return "", fmt.Errorf("unknown social platform: %s (use x, instagram, github, linkedin, youtube, tiktok, telegram, whatsapp)", platform)
		}
	default:
		return "", fmt.Errorf("unknown content type: %s", req.Type)
	}
}

func urlQueryEscape(s string) string {
	var b strings.Builder
	for _, r := range []byte(s) {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' || r == '~' {
			b.WriteByte(r)
		} else {
			fmt.Fprintf(&b, "%%%02X", r)
		}
	}
	return b.String()
}

func isFloat(s string) bool {
	dot := false
	for i, r := range s {
		switch {
		case r >= '0' && r <= '9':
		case r == '-' && i == 0:
		case r == '.' && !dot:
			dot = true
		default:
			return false
		}
	}
	return len(s) > 0
}

func parseEventTime(v string) (string, error) {
	v = strings.TrimSpace(v)
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, v); err == nil {
			return t.UTC().Format("20060102T150405Z"), nil
		}
	}
	return "", fmt.Errorf("unrecognized datetime %q", v)
}
