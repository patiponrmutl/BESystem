package handlers

import "strconv"

// แปลง string -> int; ถ้าแปลงไม่ได้ให้คืนค่าเริ่มต้น
func atoiOr(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func fmtUint(u uint) string {
	return strconv.FormatUint(uint64(u), 10)
}
