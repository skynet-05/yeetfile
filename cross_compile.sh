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

projects=(
    "cli"
    "backend")

VER="$(go run ./utils/print_version.go)"

for project in "${projects[@]}"
do
    for platform in "${platforms[@]}"
    do
        echo "Compiling $project for $platform..."
        platform_split=(${platform//\// })
        GOOS=${platform_split[0]}
        GOARCH=${platform_split[1]}

        if [ $project = "cli" ]; then
            output_name="yeetfile"
        else
            output_name="yeetfile-server"
        fi

        tar_name="${output_name}_${GOOS}_${GOARCH}_${VER}.tar.gz"
        if [ $GOOS = "darwin" ]; then
            tar_name="${output_name}_macos_${GOARCH}_${VER}.tar.gz"
        elif [ $GOARCH = "arm" ]; then
            tar_name="${output_name}-${GOOS}-arm32-${VER}.tar.gz"
        fi

        if [ $GOOS = "windows" ]; then
            output_name+=".exe"
        fi

        compile_cmd="GOOS=$GOOS GOARCH=$GOARCH go build -ldflags='-s -w' -o $output_name ./$project"
        echo "└ $compile_cmd"
        eval $compile_cmd
        if [ $? -ne 0 ]; then
            echo "An error has occurred! Aborting the script execution..."
            exit 1
        fi

        tar -czvf out/$tar_name $output_name
        rm -f $output_name
    done
done
