package utils

import (
	"net/url"
	"path"
)

func GetFilenameFromUri(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	return path.Base(u.Path), nil
}
