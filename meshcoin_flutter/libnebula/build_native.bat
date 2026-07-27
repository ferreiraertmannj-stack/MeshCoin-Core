@echo off
echo Compilando libnebula em Go para Windows (Bare Metal)...
go build -buildmode=c-shared -o libnebula.dll pqc_neonhash.go
echo.
echo Para compilar para Android, use o NDK cruzado:
echo set GOOS=android
echo set GOARCH=arm64
echo set CGO_ENABLED=1
echo set CC=NDK_PATH/toolchains/llvm/prebuilt/windows-x86_64/bin/aarch64-linux-android30-clang
echo go build -buildmode=c-shared -o libnebula.so pqc_neonhash.go
echo.
echo Concluido.
