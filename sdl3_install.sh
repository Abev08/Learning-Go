# Downloads and installs SDL3, SDL3_TTF and SDL3_Image on debian based distribution

mkdir tmp

# libsdl3 - 3.2.10
echo "Downloading and installing 'libsdl3'"
url=http://launchpadlibrarian.net/790680738/libsdl3-0_3.2.10+ds-1_amd64.deb
wget -q $url -O tmp/libsdl3-0.deb
sudo apt -q install ./tmp/libsdl3-0.deb
echo ""

# libsdl3-ttf - 3.2.2
echo "Downloading and installing 'libsdl3-ttf'"
url=http://launchpadlibrarian.net/790679301/libsdl3-ttf0_3.2.2+ds-1_amd64.deb
wget -q $url -O tmp/libsdl3-ttf0.deb
sudo apt -q install ./tmp/libsdl3-ttf0.deb
echo ""

# libsdl3-image - 3.2.4
echo "Downloading and installing 'libsdl3-image' dependencies"
url=http://launchpadlibrarian.net/782082495/libpng16-16t64_1.6.47-1.1_amd64.deb # libpng16 - 1.6.47
wget -q $url -O tmp/libpng16.deb
url=http://launchpadlibrarian.net/769677986/libsharpyuv0_1.5.0-0.1_amd64.deb # libsharpyuv0 - 1.5.0
wget -q $url -O tmp/libsharpyuv0.deb
url=http://launchpadlibrarian.net/769677987/libwebp7_1.5.0-0.1_amd64.deb # libwebp7 - 1.5.0
wget -q $url -O tmp/libwebp7.deb
url=http://launchpadlibrarian.net/769677989/libwebpdemux2_1.5.0-0.1_amd64.deb # libwebpdemux2 - 1.5.0
wget -q $url -O tmp/libwebpdemux2.deb
sudo apt -q install ./tmp/libpng16.deb ./tmp/libsharpyuv0.deb ./tmp/libwebp7.deb ./tmp/libwebpdemux2.deb
echo "Downloading and installing 'libsdl3-image'"
url=http://launchpadlibrarian.net/780310966/libsdl3-image0_3.2.4+ds-1_amd64.deb # libsdl3-image0 - 3.2.4
wget -q $url -O tmp/libsdl3-image0.deb
sudo apt -q install ./tmp/libsdl3-image0.deb

rm -r tmp
