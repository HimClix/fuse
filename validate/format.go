package validate

import (
	"encoding/base64"
	"encoding/json"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"
)

func registerFormatValidators(e *Engine) {
	e.validators["email"] = validateEmail
	e.validators["url"] = validateURL
	e.validators["http_url"] = validateHTTPURL
	e.validators["uri"] = validateURI
	e.validators["ip"] = validateIP
	e.validators["ipv4"] = validateIPv4
	e.validators["ipv6"] = validateIPv6
	e.validators["cidr"] = validateCIDR
	e.validators["mac"] = validateMAC
	e.validators["hostname"] = validateHostname
	e.validators["fqdn"] = validateFQDN
	e.validators["hostname_port"] = validateHostnamePort
	e.validators["uuid"] = validateUUID
	e.validators["uuid4"] = validateUUID4
	e.validators["json"] = validateJSON
	e.validators["jwt"] = validateJWT
	e.validators["base64"] = validateBase64
	e.validators["semver"] = validateSemver
	e.validators["datetime"] = validateDatetime
	e.validators["timezone"] = validateTimezone
	e.validators["cron"] = validateCron
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func validateEmail(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok || s == "" {
		return false
	}
	return emailRegex.MatchString(s)
}

func validateURL(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok || s == "" {
		return false
	}
	u, err := url.ParseRequestURI(s)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func validateHTTPURL(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok || s == "" {
		return false
	}
	u, err := url.ParseRequestURI(s)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func validateURI(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok || s == "" {
		return false
	}
	_, err := url.ParseRequestURI(s)
	return err == nil
}

func validateIP(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok {
		return false
	}
	return net.ParseIP(s) != nil
}

func validateIPv4(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok {
		return false
	}
	ip := net.ParseIP(s)
	return ip != nil && ip.To4() != nil
}

func validateIPv6(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok {
		return false
	}
	ip := net.ParseIP(s)
	return ip != nil && ip.To4() == nil
}

func validateCIDR(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok {
		return false
	}
	_, _, err := net.ParseCIDR(s)
	return err == nil
}

func validateMAC(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok {
		return false
	}
	_, err := net.ParseMAC(s)
	return err == nil
}

var hostnameRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)

func validateHostname(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok || s == "" || len(s) > 253 {
		return false
	}
	return hostnameRegex.MatchString(s)
}

func validateFQDN(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok || s == "" {
		return false
	}
	if s[len(s)-1] == '.' {
		s = s[:len(s)-1]
	}
	if !strings.Contains(s, ".") {
		return false
	}
	return hostnameRegex.MatchString(s)
}

func validateHostnamePort(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok || s == "" {
		return false
	}
	host, port, err := net.SplitHostPort(s)
	if err != nil || host == "" || port == "" {
		return false
	}
	return true
}

var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func validateUUID(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok {
		return false
	}
	return uuidRegex.MatchString(s)
}

func validateUUID4(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok || !uuidRegex.MatchString(s) {
		return false
	}
	return s[14] == '4' && (s[19] == '8' || s[19] == '9' || s[19] == 'a' || s[19] == 'b' ||
		s[19] == 'A' || s[19] == 'B')
}

func validateJSON(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok {
		return false
	}
	return json.Valid([]byte(s))
}

func validateJWT(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok || s == "" {
		return false
	}
	parts := strings.Split(s, ".")
	return len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != ""
}

func validateBase64(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok || s == "" {
		return false
	}
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

var semverRegex = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(-[a-zA-Z0-9.]+)?(\+[a-zA-Z0-9.]+)?$`)

func validateSemver(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok {
		return false
	}
	return semverRegex.MatchString(s)
}

func validateDatetime(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok || s == "" {
		return false
	}
	layout := info.Param
	if layout == "" {
		layout = time.RFC3339
	}
	_, err := time.Parse(layout, s)
	return err == nil
}

func validateTimezone(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok || s == "" || s == "Local" {
		return false
	}
	_, err := time.LoadLocation(s)
	return err == nil
}

var cronRegex = regexp.MustCompile(`^(\S+\s+){4}\S+$`)

func validateCron(info FieldInfo) bool {
	s, ok := info.Value.(string)
	if !ok || s == "" {
		return false
	}
	return cronRegex.MatchString(strings.TrimSpace(s))
}
