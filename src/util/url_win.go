// +build windows

package util

import (
	"dxkite.cn/go-storage/src/common"
	"fmt"
	"golang.org/x/sys/windows/registry"
	"strconv"
)

// 注册URL打开
// 参考：https://docs.microsoft.com/en-us/previous-versions/windows/internet-explorer/ie-developer/platform-apis/aa767914(v=vs.85)
func registerURLProtocol(proto, name, icon, cmd string) error {
	// 注册协议基本
	if k, _, err := registry.CreateKey(registry.CLASSES_ROOT, proto, registry.ALL_ACCESS); err != nil {
		return err
	} else {
		if er := k.SetStringValue("", name); er != nil {
			return er
		}
		if er := k.SetStringValue("URL Protocol", ""); er != nil {
			return er
		}
	}
	// 设置图标
	if k, _, err := registry.CreateKey(registry.CLASSES_ROOT, proto+`\DefaultIcon`, registry.ALL_ACCESS); err != nil {
		return err
	} else {
		if er := k.SetStringValue("", icon); er != nil {
			return er
		}
	}
	// 设置打开命令
	if k, _, err := registry.CreateKey(registry.CLASSES_ROOT, proto+`\shell\open\command`, registry.ALL_ACCESS); err != nil {
		return err
	} else {
		if er := k.SetStringValue("", cmd); er != nil {
			return er
		}
	}
	return nil
}

// 注册文件关联
// 参考：https://docs.microsoft.com/en-us/windows/win32/shell/fa-file-types
func registerFileAssociate(ext, icon, cmd, name, info string) error {
	// 注册协议基本
	if k, _, err := registry.CreateKey(registry.CLASSES_ROOT, name, registry.ALL_ACCESS); err != nil {
		return err
	} else {
		if er := k.SetStringValue("", info); er != nil {
			return er
		}
	}
	// 设置图标
	if k, _, err := registry.CreateKey(registry.CLASSES_ROOT, name+`\DefaultIcon`, registry.ALL_ACCESS); err != nil {
		return err
	} else {
		if er := k.SetStringValue("", icon); er != nil {
			return er
		}
	}
	// 设置打开命令
	if k, _, err := registry.CreateKey(registry.CLASSES_ROOT, name+`\shell\open\command`, registry.ALL_ACCESS); err != nil {
		return err
	} else {
		if er := k.SetStringValue("", cmd); er != nil {
			return er
		}
	}
	// 设置打开命令
	if k, _, err := registry.CreateKey(registry.CLASSES_ROOT, ext, registry.ALL_ACCESS); err != nil {
		return err
	} else {
		if er := k.SetStringValue("", name); er != nil {
			return er
		}
	}
	return nil
}

func Install(exec string) error {
	// 检测图标
	icon := strconv.Quote(exec) + `,0`
	if common.FileExist(exec + ".ico") {
		icon = strconv.Quote(exec + ".ico")
	}
	if er := registerURLProtocol(common.BASE_PROTOCOL, "Go Storage", icon, strconv.Quote(exec)+` "%1"`); er != nil {
		return er
	}
	if er := registerFileAssociate(common.EXT_META, icon, strconv.Quote(exec)+` "%1"`, common.META_NAME, common.META_INFO); er != nil {
		return er
	}
	return nil
}

type MsgError struct {
	msg string
	err error
}

func (ue MsgError) Error() string {
	return fmt.Sprintf("%s %v", ue.msg, ue.err)
}

func deleteURLProtocol(name string) error {
	if err := registry.DeleteKey(registry.CLASSES_ROOT, name+`\DefaultIcon`); err != nil && err != registry.ErrNotExist {
		return MsgError{"delete reg:" + common.BASE_PROTOCOL + `\DefaultIcon`, err}
	}

	if err := registry.DeleteKey(registry.CLASSES_ROOT, name+`\shell\open\command`); err != nil && err != registry.ErrNotExist {
		return MsgError{"delete reg:" + name + `\shell\open\command`, err}
	}

	if err := registry.DeleteKey(registry.CLASSES_ROOT, name+`\shell\open`); err != nil && err != registry.ErrNotExist {
		return MsgError{"delete reg:" + name + `\shell\open`, err}
	}

	if err := registry.DeleteKey(registry.CLASSES_ROOT, name+`\shell`); err != nil && err != registry.ErrNotExist {
		return MsgError{"delete reg:" + name + `\shell`, err}
	}

	if err := registry.DeleteKey(registry.CLASSES_ROOT, name); err != nil && err != registry.ErrNotExist {
		return MsgError{"delete reg:" + name, err}
	}

	return nil
}

func Uninstall(path string) error {
	if err := registry.DeleteKey(registry.CLASSES_ROOT, common.EXT_META); err != nil && err != registry.ErrNotExist {
		return MsgError{"delete reg:" + common.EXT_META, err}
	}
	if err := deleteURLProtocol(common.BASE_PROTOCOL); err != nil {
		return err
	}
	if err := deleteURLProtocol(common.META_NAME); err != nil {
		return err
	}
	return nil
}
