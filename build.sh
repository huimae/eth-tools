xgo --targets=windows/amd64 --ldflags="-H=windowsgui -extldflags=-s" --pkg github.com/naiba/hunter -v . &&
    xgo --targets=darwin/amd64 --ldflags="-extldflags=-s" --pkg github.com/naiba/hunter -v .
