// +build !windows

package util

import "fmt"

func Install(exec string) error {
	fmt.Sprintf("only support windows")
	return nil
}

func Uninstall(path string) {
	fmt.Sprintf("only support windows")
	return nil
}
