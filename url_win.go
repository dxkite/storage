// +build windows

package storage

import (
	"fmt"
	"golang.org/x/sys/windows/registry"
	"strconv"
)

type RegError struct {
	opt string
	reg string
	err error
}

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

func InstallURL(exec string) error {
	// 检测图标
	icon := strconv.Quote(exec) + `,0`
	if FileExist(exec + ".ico") {
		icon = strconv.Quote(exec + ".ico")
	}
	if er := registerURLProtocol(BASE_PROTOCOL, APP_NAME, icon, strconv.Quote(exec)+` "%1"`); er != nil {
		return er
	}
	if er := registerFileAssociate(EXT_META, icon, strconv.Quote(exec)+` "%1"`, META_NAME, META_INFO); er != nil {
		return er
	}
	return nil
}

func (ue RegError) Error() string {
	return fmt.Sprintf("%s `%s` error: %v", ue.opt, ue.reg, ue.err)
}

func deleteURLProtocol(name string) error {
	if err := registry.DeleteKey(registry.CLASSES_ROOT, name+`\DefaultIcon`); err != nil && err != registry.ErrNotExist {
		return RegError{"delete", BASE_PROTOCOL + `\DefaultIcon`, err}
	}

	if err := registry.DeleteKey(registry.CLASSES_ROOT, name+`\shell\open\command`); err != nil && err != registry.ErrNotExist {
		return RegError{"delete", name + `\shell\open\command`, err}
	}

	if err := registry.DeleteKey(registry.CLASSES_ROOT, name+`\shell\open`); err != nil && err != registry.ErrNotExist {
		return RegError{"delete", name + `\shell\open`, err}
	}

	if err := registry.DeleteKey(registry.CLASSES_ROOT, name+`\shell`); err != nil && err != registry.ErrNotExist {
		return RegError{"delete", name + `\shell`, err}
	}

	if err := registry.DeleteKey(registry.CLASSES_ROOT, name); err != nil && err != registry.ErrNotExist {
		return RegError{"delete", name, err}
	}

	return nil
}

func UninstallURL(path string) error {
	if err := registry.DeleteKey(registry.CLASSES_ROOT, EXT_META); err != nil && err != registry.ErrNotExist {
		return RegError{"delete", EXT_META, err}
	}
	if err := deleteURLProtocol(BASE_PROTOCOL); err != nil {
		return err
	}
	if err := deleteURLProtocol(META_NAME); err != nil {
		return err
	}
	return nil
}
