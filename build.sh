rm -rf eth-* && rm -rf candy-* && xgo --targets=windows/amd64 --ldflags="-s -w -H=windowsgui" --pkg github.com/naiba/eth-tools/cmd/$1 -v . &&
    xgo --targets=darwin/amd64 --ldflags="-s -w" --pkg github.com/naiba/eth-tools/cmd/$1 -v .
