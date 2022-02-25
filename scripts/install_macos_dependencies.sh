#!/bin/bash

echo "Preparing dependencies for macos"

brew install gmp
brew install openssl@1.1
sudo ln -sf /opt/homebrew/opt/openssl@1.1/include/openssl /usr/local/include/openssl
sudo mkdir -p /usr/local/opt/openssl
sudo ln -sf /opt/homebrew/opt/openssl@1.1/lib /usr/local/opt/openssl/lib
sudo mkdir -p /usr/local/opt/gmp
sudo ln -sf /opt/homebrew/opt/gmp/include /usr/local/include/gmp
sudo ln -sf /opt/homebrew/opt/gmp/lib /usr/local/opt/gmp/lib

# Prepare temp directory
mkdir -p temp
pushd temp

# Prepare harmony dependencies
git clone --depth=1 https://github.com/harmony-one/bls.git
pushd bls
git checkout 2b7e49894c0f15f5c40cf74046505b7f74946e52
popd

git clone --depth=1 https://github.com/harmony-one/mcl.git
pushd mcl
git checkout 99e9aa76e84415e753956c618cbc662b2f373df1
popd

export CGO_CFLAGS="-I$PWD/bls/include -I$PWD/mcl/include -I/usr/local/include"
export CGO_LDFLAGS="-L$PWD/bls/lib -L/usr/local/opt/openssl/lib"
export LD_LIBRARY_PATH="$PWD/bls/lib:$PWD/mcl/lib:/usr/local/opt/openssl/lib:/usr/local/opt/gmp/lib"
export LIBRARY_PATH=$LD_LIBRARY_PATH
export DYLD_FALLBACK_LIBRARY_PATH=$LD_LIBRARY_PATH

echo "making mcl"
make -C $PWD/mcl -j8
echo "making bls"
make -C $PWD/bls BLS_SWAP_G=1 -j8

sudo ln -sf $PWD/bls/lib/libbls384_256.dylib /usr/local/lib/libbls384_256.dylib
sudo ln -sf $PWD/bls/lib/libmcl.dylib /usr/local/lib/libmcl.dylib
