package helper

import "time"

func ParseTime(raw string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,          // 2025-08-17T18:30:00Z
		"2006-01-02",          // 2025-08-17
		"2006-01-02 15:04:05", // 2025-08-17 18:30:00
	}
	var err error
	for _, layout := range layouts {
		t, e := time.Parse(layout, raw)
		if e == nil {
			return t, nil
		}
		err = e
	}
	return time.Time{}, err
}
