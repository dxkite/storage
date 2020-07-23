// +build !windows

package storage

import "fmt"

func InstallURL(exec string) error {
	fmt.Sprintf("only support windows")
	return nil
}

func UninstallURL(path string) {
	fmt.Sprintf("only support windows")
	return nil
}
