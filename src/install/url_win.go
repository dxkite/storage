// +build windows

package install

import (
	"dxkite.cn/go-storage/src/config"
	"golang.org/x/sys/windows/registry"
	"log"
	"strconv"
)

// 注册URL打开
// 参考：https://docs.microsoft.com/en-us/previous-versions/windows/internet-explorer/ie-developer/platform-apis/aa767914(v=vs.85)
func RegisterURLProtocol(proto, name, icon, cmd string) error {
	// 注册协议基本
	if k, exist, err := registry.CreateKey(registry.CLASSES_ROOT, proto, registry.ALL_ACCESS); err != nil {
		return err
	} else {
		if exist {
			log.Println("exist protocol register")
			return nil
		}
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
func RegisterFileAssociate(ext, icon, cmd, name, info string) error {
	// 注册协议基本
	if k, exist, err := registry.CreateKey(registry.CLASSES_ROOT, name, registry.ALL_ACCESS); err != nil {
		return err
	} else {
		if exist {
			log.Println("exist associate register")
			return nil
		}
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

func CreateHelper(exec string) error {
	if er := RegisterURLProtocol(config.BASE_PROTOCOL, "Go Storage", strconv.Quote(exec)+`,0`, strconv.Quote(exec)+` -meta "%1"`); er != nil {
		return er
	}
	if er := RegisterFileAssociate(".meta", strconv.Quote(exec)+`,0`, strconv.Quote(exec)+` -meta "%1"`, "GoStorageMetaFile", "Go Storage Download Meta File"); er != nil {
		return er
	}
	return nil
}
