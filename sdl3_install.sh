# Downloads and installs SDL3, SDL3_TTF and SDL3_Image on debian based distribution
# Works on Ubuntu LTS 24.0x, made to be used with WSL.

mkdir tmp

# https://launchpad.net/ubuntu/+source/libsdl3 - 3.2.24
echo "Downloading and installing 'libsdl3'"
url=http://launchpadlibrarian.net/827297829/libsdl3-0_3.2.24+ds-1_amd64.deb
wget -q $url -O tmp/libsdl3-0.deb
sudo apt -q install ./tmp/libsdl3-0.deb
echo ""

# https://launchpad.net/ubuntu/+source/libsdl3-ttf - 3.2.2
echo "Downloading and installing 'libsdl3-ttf'"
url=http://launchpadlibrarian.net/790679301/libsdl3-ttf0_3.2.2+ds-1_amd64.deb
wget -q $url -O tmp/libsdl3-ttf0.deb
sudo apt -q install ./tmp/libsdl3-ttf0.deb
echo ""

# https://launchpad.net/ubuntu/+source/libsdl3-image - 3.2.4
echo "Downloading and installing 'libsdl3-image' dependencies"
url=http://launchpadlibrarian.net/810584546/libpng16-16t64_1.6.50-1_amd64.deb # https://launchpad.net/ubuntu/+source/libpng1.6 - 1.6.50
wget -q $url -O tmp/libpng16.deb
url=http://launchpadlibrarian.net/769677986/libsharpyuv0_1.5.0-0.1_amd64.deb # https://launchpad.net/ubuntu/plucky/amd64/libsharpyuv0 - 1.5.0
wget -q $url -O tmp/libsharpyuv0.deb
url=http://launchpadlibrarian.net/769677987/libwebp7_1.5.0-0.1_amd64.deb # https://launchpad.net/ubuntu/+source/libwebp - 1.5.0
wget -q $url -O tmp/libwebp7.deb
url=http://launchpadlibrarian.net/769677989/libwebpdemux2_1.5.0-0.1_amd64.deb # https://launchpad.net/ubuntu/plucky/amd64/libwebpdemux2 - 1.5.0
wget -q $url -O tmp/libwebpdemux2.deb
sudo apt -q install ./tmp/libpng16.deb ./tmp/libsharpyuv0.deb ./tmp/libwebp7.deb ./tmp/libwebpdemux2.deb
echo "Downloading and installing 'libsdl3-image'"
url=http://launchpadlibrarian.net/780310966/libsdl3-image0_3.2.4+ds-1_amd64.deb # https://launchpad.net/ubuntu/plucky/amd64/libsdl3-image0 - 3.2.4
wget -q $url -O tmp/libsdl3-image0.deb
sudo apt -q install ./tmp/libsdl3-image0.deb

rm -r tmp
