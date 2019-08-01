rm -rf eth-debugger-* && xgo --targets=windows/amd64 --ldflags="-s -w -H=windowsgui" --pkg github.com/naiba/eth-debugger -v . &&
    xgo --targets=darwin/amd64 --ldflags="-s -w" --pkg github.com/naiba/eth-debugger -v .
