# TuneDb installation on Windows

Get "TunedbWindows.zip" from https://github.com/py60800/tunedb/Windows

Extract (unzip) the files in your working space.

## Install MuseScore and Python

Install MuseScore4 (=> https://musescore.org)

Check that MuseScore is available from the following  path :
       'C:\Program Files\MuseScore 4\bin\MuseScore4.exe'
if not, proper path will have to be adjusted in the "config.yml" file
 that is created during the first run.

Install Python (=> https://www.python.org/downloads/windows/)

Verify python installation and install required module :
- open Windows terminal (cmd)
- run command "python -V" to check if python is properly installed
- run command "pip install pyparsing"

## Prepare your environment

### Locate mp3 files
The default directory where TuneDb will search MP3 file is "%HOMEPATH%/Music/mp3".
TuneDb will search within subdirectories.
Additionnal repositories can be added later from TuneDB (Menu Config) 

### Locate tune files
The default directory where TuneDb will install tune files is "%HOMEPATH%/Music/MuseScore", using a subdirectory per tune kind.
Some samples are provided. The samples.zip file can be extracted in "%HOME%\Music\MuseScore" directory (optional)

## Start the application
launch "tunedb.exe" from installation directory

**Issue** Sometimes TuneDb hangs at startup on Windows, quit the app when asked by Windows and restart (if you have any clue, you're welcome...)

**Don't forget to read the Quick Start Manual (Manual.pdf) to understand how to use TuneDb**

