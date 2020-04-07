// +build !windows

package install

import "fmt"

func CreateHelper(exec string) error {
	fmt.Sprintf("only support windows")
	return nil
}
