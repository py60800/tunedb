# Windows Installation

## Install MuseScore4
MuseScore can be found [Here](https://musescore.org)

MuseScore is expected to be found at "c:\\Program Files\\MuseScore 4\\bin\\MuseScore4.exe". 
If not, "config.yml" will have to be updated after the first launch.

## Install Python 3
Python can be installed from Microsoft repositories (search Python using Windows tool bar)


Python is expected to be located at "${HOME}\\AppData\\Local\\Microsoft\\WindowsApps\\python3.exe"
If not, "config.yml" will have to be updated after the first launch.

Download required package for Python


   
    pip install pyparsing

## Install MSYS 2
MSYS is available here [MSYS 2](https://www.msys2.org/)

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

Build

    git clone github.com/py60800/tunedb
    cd tunedb
    go build
 
 Source code can also be available as a zip from Github (visit [Github](https://github.com/py60800/tunedb))

It may be required to "go get" some module (check error messages)

When the build succeeds, the application is available as "tunedb.exe".

## Prepare your environnement before the first launch:

**locate your MP3 files:**
By default, tunedb will search for all MP3s in the folder ~/Music/mp3 (including subfolders).
Additional local repositories can be added later using the Config menu.

Note that usual MP3 tags are used to index MP3 files (Artist, Album, Title). Sample rates other than 48000 or 44100 Hz may not work.

**Prepare the repository for your tunes:**

The default structure is $HOME/Music/MuseScore/_TuneKind_/  with one subfolder per tune kind

You can use the sample.zip (from the samples directory) to create your initial environment.
Extract it to $HOME/Music/MuseScore folder)

### Start the application

Run "tunedb.exe" from MSYS2/UCRT64 prompt.

The launch of TuneDB can scripted to ease the launch from Windows native environment.



