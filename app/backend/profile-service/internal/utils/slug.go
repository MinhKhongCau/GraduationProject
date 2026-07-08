// File: internal/utils/slug.go
package utils

import (
	"strings"

	"github.com/google/uuid"
)

var vietnameseDiacritics = map[rune]rune{
	'à': 'a', 'á': 'a', 'ả': 'a', 'ã': 'a', 'ạ': 'a',
	'ă': 'a', 'ằ': 'a', 'ắ': 'a', 'ẳ': 'a', 'ẵ': 'a', 'ặ': 'a',
	'â': 'a', 'ầ': 'a', 'ấ': 'a', 'ẩ': 'a', 'ẫ': 'a', 'ậ': 'a',
	'đ': 'd',
	'è': 'e', 'é': 'e', 'ẻ': 'e', 'ẽ': 'e', 'ẹ': 'e',
	'ê': 'e', 'ề': 'e', 'ế': 'e', 'ể': 'e', 'ễ': 'e', 'ệ': 'e',
	'ì': 'i', 'í': 'i', 'ỉ': 'i', 'ĩ': 'i', 'ị': 'i',
	'ò': 'o', 'ó': 'o', 'ỏ': 'o', 'õ': 'o', 'ọ': 'o',
	'ô': 'o', 'ồ': 'o', 'ố': 'o', 'ổ': 'o', 'ỗ': 'o', 'ộ': 'o',
	'ơ': 'o', 'ờ': 'o', 'ớ': 'o', 'ở': 'o', 'ỡ': 'o', 'ợ': 'o',
	'ù': 'u', 'ú': 'u', 'ủ': 'u', 'ũ': 'u', 'ụ': 'u',
	'ư': 'u', 'ừ': 'u', 'ứ': 'u', 'ử': 'u', 'ữ': 'u', 'ự': 'u',
	'ỳ': 'y', 'ý': 'y', 'ỷ': 'y', 'ỹ': 'y', 'ỵ': 'y',
}

func stripDiacritics(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if replacement, ok := vietnameseDiacritics[r]; ok {
			b.WriteRune(replacement)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Slugify chuyển một chuỗi (tên người dùng) thành dạng slug URL-safe: chữ thường,
// bỏ dấu tiếng Việt, khoảng trắng/ký tự đặc biệt thay bằng dấu gạch ngang.
func Slugify(s string) string {
	stripped := stripDiacritics(s)

	var b strings.Builder
	lastDash := false
	for _, r := range stripped {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash && b.Len() > 0 {
				b.WriteRune('-')
				lastDash = true
			}
		}
	}

	slug := strings.Trim(b.String(), "-")
	if slug == "" {
		slug = "profile"
	}
	return slug
}

// GenerateUniqueSlug tạo slug từ tên kèm hậu tố ngẫu nhiên ngắn để đảm bảo tính duy nhất
// (uniqueIndex trên cột slug), phòng trường hợp trùng tên.
func GenerateUniqueSlug(name string) string {
	base := Slugify(name)
	suffix := strings.ReplaceAll(uuid.NewString(), "-", "")[:8]
	return base + "-" + suffix
}
