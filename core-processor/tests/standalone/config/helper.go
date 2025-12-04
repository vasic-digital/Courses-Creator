package config_test

func maskAPIKey(key string) string {
	if key == "" {
		return "[not set]"
	}
	if len(key) <= 8 {
		return "[masked]"
	}
	return key[:4] + "[...]" + key[len(key)-4:]
}
