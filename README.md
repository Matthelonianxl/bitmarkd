# bitmarkd - Main program

[![GoDoc](https://godoc.org/github.com/bitmark-inc/bitmarkd?status.svg)](https://godoc.org/github.com/bitmark-inc/bitmarkd)

Prerequisites

* Install the go language package for your system
* Configure environment variables for go system
* install the ZMQ4 and Argon2 libraries

For shell add the following to the shell's profile
(remark the `export CC=clang` if you wish to use gcc)
~~~~~
# check for go installation
GOPATH="${HOME}/gocode"
if [ -d "${GOPATH}" ]
then
  gobin="${GOPATH}/bin"
  export GOPATH
  export PATH="${PATH}:${gobin}"
  # needed for FreeBSD 10 and later
  export CC=clang
else
  unset GOPATH
fi
unset gobin
~~~~~

On FreeBSD

~~~~~
pkg install libzmq4 libargon2
~~~~~

On  MacOSX
(be sure that homebrew is installed correctly)
~~~~
brew install argon2

brew tap bitmark-inc/bitmark
brew install zeromq41
~~~~

On Ubuntu
(tested dor distribution 18.04)

Install following packages
   `sudo apt install libargon2-0-dev uuid-dev libzmq3-dev`

To compile simply:

~~~~~
go get github.com/bitmark-inc/bitmarkd
go install -v github.com/bitmark-inc/bitmarkd/command/bitmarkd
~~~~~

# Set up

Create the configuration directory, copy sample configuration, edit it to
set up IPs, ports and local bitcoin testnet connection.

~~~~~
mkdir -p ~/.config/bitmarkd
cp command/bitmarkd/bitmarkd.conf.sample  ~/.config/bitmarkd/bitmarkd.conf
${EDITOR}   ~/.config/bitmarkd/bitmarkd.conf
~~~~~

To see the bitmarkd sub-commands:

~~~~~
bitmarkd --config-file="${HOME}/.config/bitmarkd/bitmarkd.conf" help
~~~~~

Generate key files and certificates.

~~~~~
bitmarkd --config-file="${HOME}/.config/bitmarkd/bitmarkd.conf" gen-peer-identity
bitmarkd --config-file="${HOME}/.config/bitmarkd/bitmarkd.conf" gen-rpc-cert
bitmarkd --config-file="${HOME}/.config/bitmarkd/bitmarkd.conf" gen-proof-identity
~~~~~

Start the program.

~~~~~
bitmarkd --config-file="${HOME}/.config/bitmarkd/bitmarkd.conf" start
~~~~~

Note that a similar process is needed for the prooferd (mining subsystem)

# Prebuilt Binary

* Flatpak

    Please refer to [wiki](https://github.com/bitmark-inc/bitmarkd/wiki/Instruction-for-Flatpak-Prebuilt)

* Docker

    Please refer to [bitmark-node](https://github.com/bitmark-inc/bitmark-node)

# Coding

* setup git hooks
  
  Link git hooks directory, run command `./scripts/setup-hook.sh` at root of bitmarkd 
  directory. Currently it provides checkings for two stages:
  
  1. Before commit (`pre-commt`)

	Runs `go lint` for every modified file(s). It shows suggestions but not
    necessary to follow. 

  2. Before push to remote (`pre-push`)
  
  	Runs `go test` for whole directory except `vendor` one. It is
    mandatory to pass this check because generally, new modifications should not
    break existing logic/behavior.
    
    Other optional actions are `sonaqube` and `go tool vet`. These two are
    optional to follow since static code analysis just provide some advice.
  
* all variables are camel case i.e. no underscores
* labels are all lowercase with '_' between words
* imports and one single block
* all break/continue must have label
* avoid break in switch and select
