package upload

import (
	"testing"
)

func TestPluginUploader_Upload(t *testing.T) {
	uploadTest(t, "plugin://cmd?exec=python&args=./testdata/plugin-vim-cn.py")
}
