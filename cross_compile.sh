#!/bin/sh

dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)

mkdir -p $dir/out/
rm -rf $dir/out/*

platforms=(
    "windows/386"
    "windows/amd64"
    "windows/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "linux/arm"
    "linux/amd64"
    "linux/arm64"
    "linux/386")

VER="$(go run ./utils/print_version.go)"

for platform in "${platforms[@]}"
do
    echo "Compiling for $platform..."
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}

    output_name="yeetfile"
    zip_name="yeetfile-$GOOS-$GOARCH-$VER.zip"
    if [ $GOOS = "darwin" ]; then
        zip_name="yeetfile-macos-$GOARCH-$VER.zip"
    elif [ $GOARCH = "arm" ]; then
        zip_name="yeetfile-$GOOS-arm32-$VER.zip"
    fi

    if [ $GOOS = "windows" ]; then
        output_name+=".exe"
    fi

    GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o $output_name ./cli
    if [ $? -ne 0 ]; then
        echo "An error has occurred! Aborting the script execution..."
        exit 1
    fi

    zip out/$zip_name $output_name
    rm -f $output_name
done

