package utils

import (
	. "os"
)

func Isfileexist(file string) bool {
	fi, err := Stat(file)
	if err != nil {
		if IsNotExist(err) {
			return false
		}
		return false
	}

	return !fi.IsDir()
}

func Isdirectoryexist(dname string) bool {
	fi, err := Stat(dname)
	if err != nil {
		if IsNotExist(err) {
			return false
		}
		return false
	}

	return fi.IsDir()
}

func Isfdexist(fd string) bool {
	_, err := Stat(fd)
	if err != nil {
		if IsNotExist(err) {
			return false
		}
		return false
	}

	return true
}
