#!/usr/bin/env python
import requests
import sys

if __name__ == "__main__":
    resp = requests.post('https://img.vim-cn.com/', files = [
        ('file', sys.stdin.buffer.raw.read())
    ])
    print(resp.text.strip())