#!/bin/bash

case $(uname | tr '[:upper:]' '[:lower:]') in
  linux*)
    ./scripts/install_linux_dependencies.sh
    ;;
  darwin*)
    ./scripts/install_macos_dependencies.sh
    ;;
  *)
    echo "Unsupported os"
    ;;
esac