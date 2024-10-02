Collection of smaller projects created while learning and exploring Go.
<br><br>

#### Easiest way (in my opinion) to install GCC (MSYS2 and MinGW) on Windows:
1. Open terminal (cmd.exe) as administrator if possible
2. Run the command: ```winget install msys2.msys2```
3. Start ucrt64.exe from "C:\msys64"
4. Run the command: ```pacman -S --needed base-devel mingw-w64-ucrt-x86_64-toolchain```
5. Add ```C:\msys64\ucrt64\bin``` to system path
6. Restart the terminal and confirm that installation works with command: ```gcc --version```
<br><br>

#### To use SDL2 with Go on Windows
There are few requirements to use SDL2 with Go.  
You can follow [veandco tutorial](https://github.com/veandco/go-sdl2?tab=readme-ov-file#requirements) how to get it working.  
Or you can follow these steps:
1. Install GCC (described in the list above)  
2. Download and extract SDL2 libraries  
   - Go to https://github.com/libsdl-org/SDL/releases
   - Download SDL2-devel-[version]-mingw.zip - version should be the newest one
   - Extract the downloaded package (or use .zip explorer like 7-zip)
   - Go to ```x86_64-w64-mingw32``` folder in downloaded package
   - Copy contents of the folder into ```C:\msys64\ucrt64\```
3. Optional but highly recomended: repeat step 2 for [SDL2_ttf](https://github.com/libsdl-org/SDL_ttf/releases) and [SDL2_img](https://github.com/libsdl-org/SDL_image/releases)

Great tutorial series on creating games in Go: [Games With Go](https://www.youtube.com/watch?v=9D4yH7e_ea8&list=PLDZujg-VgQlZUy1iCqBbe5faZLMkA3g2x).
