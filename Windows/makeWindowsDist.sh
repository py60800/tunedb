
mkdir tunedb
cd tunedb
cp ../../tunedb.exe .
for t in `ldd tunedb.exe | awk '/\/ucrt64\// { print $3 }'` ; do cp $t . ; done
for t in /ucrt64/bin/librsvg-2-2.dll /ucrt64/bin/libxml2-2.dll \
       	/ucrt64/bin/libiconv-2.dll /ucrt64/bin/libcharset-1.dll /ucrt64/bin/zlib1.dll \
	; do cp $t . ; done
mkdir lib
cp -r /ucrt64/lib/gdk-pixbuf-2.0 ./lib/gdk-pixbuf-2.0
mkdir -p share/icons
cp -r /ucrt64/share/icons/* ./share/icons/
cp ../../doc/Manual.pdf .
cp ../../samples/samples.zip .
cp ../README_Windows.md .
cd ..
zip -rq TunedbWindows.zip tunedb


