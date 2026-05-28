#!/usr/bin/env sh
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
MOBILE_DIR="$ROOT_DIR/apps/mobile"
XCODE_DEVELOPER_DIR="${XCODE_DEVELOPER_DIR:-/Applications/Xcode-26.5.0.app/Contents/Developer}"

echo "Selecting Xcode developer directory: $XCODE_DEVELOPER_DIR"
sudo xcode-select -s "$XCODE_DEVELOPER_DIR"

echo
echo "Xcode version"
xcodebuild -version

echo
echo "Flutter doctor"
flutter doctor -v

echo
echo "Refreshing Flutter mobile workspace"
cd "$MOBILE_DIR"
flutter clean
flutter pub get

echo
echo "iOS dev environment refresh complete."
