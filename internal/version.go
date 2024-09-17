package internal

var version string

func Version() string {
	if version == "" {
		return "unspecified"
	}
	return version
}
