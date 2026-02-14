package utils

import "time"

// NowJakarta mengembalikan waktu sekarang di timezone Asia/Jakarta
func NowJakarta() time.Time {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		// fallback ke UTC kalau gagal load location
		return time.Now().UTC()
	}
	return time.Now().In(loc)
}
