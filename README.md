# TuneDB
TuneDB is a desktop application  intended  initially to manage my own collection of Irish tunes (Those I play and those I want to learn). The main features are :

- Score display
- Tune annotation: (How you like it, how you play it, free comment...)
- Midi playing
- Local search engine
- MP3 playing with speed and pitch adjustment
- Tune editing (using MuseScore4 as a companion application)
- Tune import in ABC format (typically from [“thesession.org”](https://thesession.org))
- sets, lists creation
- ...
    
**Important:** TuneDB has been developed primarily for a Linux Desktop and has been compiled and adapted for Windows. 

Linux users must build and install TuneDb from the source code (procedure described hereafter)

Windows users can [download a prebuilt binary](Windows/TunedbWindows.zip) or [build it from the source](https://github.com/py60800/tunedb/blob/main/WINDOWS.md)


**Note that MP3 playing depends on your own collection of MP3 files.**

## Technical details
   
- Go application using GTK3
- sqlite3 database with gorm (ORM)
- depends on an external midi synthesizer such as  fluidsynth
- Intensive use of MuseScore4 is used for tune editing  and format conversion (=>musicxml, svg)[MuseScore](https://musescore.org/)
- Use of abc2xml (and therefore python3) for ABC import
- Use of rubberband library for speed and pitch adjustment when playing mp3 [Rubberband](https://breakfastquay.com/rubberband/)
    
Tunes are stored in musescore format (.mscz) and derived files are created for different purposes (play midi and display)

## Prerequisites (Linux)
    
- Musescore4 (that can be downloaded as a portable app from [MuseScore](https://musescore.org/). It must be "installed" and made available via the PATH with the name MuseScore4
- GTK3
- Fluidsynth
- Rubberband library
- Sound system (ie : pipewire)
- python3 (for ABC files import)
- sqlite3 library
          
Debian packages: libgtk-3-common, pipewire, rubberband-cli, libsqlite3-0, python3-pyparsing

## Installation (Linux) 

For Linux TuneDb must be build from the sources

Dowload development libraries (Debian: libgtk-3-dev librubberband-dev libasound2-dev libsqlite3-dev)

From an empty directory

    git clone https://github.com/py60800/tunedb.git
    cd tunedb
    go build


## Prepare your environnement before the first launch:

**locate your MP3 files:**
By default, tunedb will search for all MP3s in the folder ~/Music/mp3 (including subfolders).
Additional local repositories can be added later using the Config menu.

Note that usual MP3 tags are used to index MP3 files (Artist, Album, Title). Sample rates other than 48000 or 44100 Hz may not work.

**Prepare the repository for your tunes:**

The default structure is ~/Music/MuseScore/_TuneKind_/  with one subfolder per tune kind

You can use the sample.zip (from the samples) to create your initial environment (unzip from ~/Music/MuseScore folder)

## Launch TuneDB

A MIDI synthesizer must be available before the start of tunedb, it can be launched manually but it’s better to use a script. Here is an example that checks if the synthesizer is active and launch it if not:

    #!/bin/sh
    aplaymidi -l | grep -i synth 
    if [ $? -ne 0 ] ; then 
    fluidsynth -a pulseaudio -q -si /usr/share/sounds/sf2/FluidR3_GM.sf2 &
       sleep 1
    fi
    ~/bin/tunedb &

On the first run, tunedb will create a working directory to hold the database and some configurations files (that can be adapted to your needs).

The default woorking directory is ~/Music/TuneDb.

Several databases can be created using the command:

    tunedb -d AlternateDirectory

Creating alternate database is recommended to test TuneDB features.

The quck start manual is available [here](https://github.com/py60800/tunedb/blob/main/doc/Manual.pdf)
