package upload

import (
	"testing"
)

func TestPluginUploader_Upload(t *testing.T) {
	uploadTest(t, "plugin://cmd?exec=./testdata/plugin-vim-cn.exe")
}
