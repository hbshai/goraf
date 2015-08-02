[[ $# -eq 0 ]] && echo "usage: release.sh <MAJOR.MIN.PATCH>" && exit 1;

VERSION=$1

[[ ! -d "./dist" ]] && echo "Creating ./dist directory..." && mkdir ./dist

function BUILD {
    # $1=GOOS $2=GOARCH $3=VERSION
    echo "Building $1_$2..."
    TARNAME="./dist/goraf-$1_$2-v$3.tar"
    GOOS=$1 GOARCH=$2 go build
    tar -cf $TARNAME ./goraf
    tar -rf $TARNAME ./programs.json
    tar -rf $TARNAME ./public
    tar -rf $TARNAME ./LICENSE
    gzip $TARNAME
    rm ./goraf
}

BUILD linux amd64 $VERSION
BUILD darwin amd64 $VERSION

echo "Building windows_amd64..."
GOOS=windows GOARCH=amd64 go build
zip -qr "./dist/goraf-windows_amd64-v$1.zip" ./goraf.exe ./programs.json ./public ./LICENSE
rm ./goraf.exe