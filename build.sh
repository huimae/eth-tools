rm -rf eth-debugger-* && xgo --targets=windows/amd64 --ldflags="-s -w -H=windowsgui -extldflags=-static" --pkg github.com/naiba/eth-debugger -v . &&
    xgo --targets=darwin/amd64 --ldflags="-extldflags=-s" --pkg github.com/naiba/eth-debugger -v .
