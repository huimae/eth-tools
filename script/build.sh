xgo --deps=https://libsdl.org/release/SDL2-2.0.9.tar.gz \
    --targets=windows/amd64 \
    -ldflags "-lSDL2 -D_REENTRANT" \
    -tags "sdl" \
    --pkg=cmd/hunter/main.go \
    -v -x \
    .
