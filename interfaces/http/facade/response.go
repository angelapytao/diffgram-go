package facade

type apiLog struct {
	Success bool              `json:"success"`
	Error   map[string]string `json:"error"`
}

func okLog() apiLog {
	return apiLog{Success: true, Error: map[string]string{}}
}

func errLog(field, msg string) apiLog {
	return apiLog{Success: false, Error: map[string]string{field: msg}}
}
