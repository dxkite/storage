# plugin

a upload plugin use stdin to upload file, get url from stdout

## plugin example

the upload plugin, can use as usn:

- `plugin://cmd?exec=python&args=./plugin-vim-cn.py` run with python (for windows)
- `plugin://cmd?exec=./plugin-vim-cn.py` run as bash (for unix)

### example plugin code

```python
#!/usr/bin/env python
import requests
import sys

if __name__ == "__main__":
    resp = requests.post('https://img.vim-cn.com/', files = [
        ('file', sys.stdin.buffer.raw.read())
    ])
    print(resp.text.strip())
```
