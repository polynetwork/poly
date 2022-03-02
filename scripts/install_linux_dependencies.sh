#!/bin/bash

echo "Preparing dependencies for linux/ubuntu"

if [ "$(grep -Ei 'debian|buntu|mint' /etc/*release)" ]; then
   sudo apt update
   sudo apt install -y libgmp-dev  libssl-dev  make gcc g++
else
   sudo yum install glibc-static gmp-devel gmp-static openssl-libs openssl-static gcc-c++
fi

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

export CGO_CFLAGS="-I$PWD/bls/include -I$PWD/mcl/include"
export CGO_LDFLAGS="-L$PWD/bls/lib"
export LD_LIBRARY_PATH=$PWD/bls/lib:$PWD/mcl/lib
export LIBRARY_PATH=$LD_LIBRARY_PATH
export DYLD_FALLBACK_LIBRARY_PATH=$LD_LIBRARY_PATH

echo "making mcl"
make -C $PWD/mcl -j8
echo "making bls"
make -C $PWD/bls minimised_static BLS_SWAP_G=1 -j8



