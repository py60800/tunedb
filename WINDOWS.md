# Windows Installation

## Install MuseScore4
MuseScore can be found [Here](https://musescore.org)

## Install Python 3
Python can be installed from Microsoft repositories (search Python using Windows tool bar)
Download required package for Python
   
    pip install pyparse

## Install MSYS 2
MSYS is available [MSYS 2](https://www.msys2.org/)

Open MSYS UCRT64 terminal and 

### Install required packages

    pacman -S mingw-w64-ucrt-x86_64-toolchain base-devel
    pacman -S mingw-w64-x86_64-gcc
    pacman -S mingw-w64-x86_64-pkg-config
    pacman -S mingw-w64-ucrt-x86_64-rubberband
    pacman -S mingw-w64-ucrt-x86_64-gtk3
    pacman -S mingw-w64-x86_64-sqlite3
    pacman -S mingw-w64-x86_64-portmidi
    pacman -S mingw-w64-x86_64-libao
    pacman -S mingw-w64-ucrt-x86_64-go

### Build the application
Set some GO environment variables:

    export GOROOT=/ucrt64/lib/go
    export GOPATH=/ucrt64
    export CGO_ENABLED=1
    export GOBIN=~/bin

Install the application :

    go install github.com/py60800/tunedb@latest
    
If the build succeeds, the application should be available in the ~/bin directory.

### Start the application

type ~/bin/tunedb to start the application.



