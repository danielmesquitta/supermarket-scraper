package time

import "time"

func SetServerTimeZone() {
	time.Local = time.UTC
}
