Collection of smaller projects created while learning and exploring Go.
<br><br>

### Easiest way (in my opinion) to install GCC (MSYS2 and MinGW) on Windows:
1. Open terminal (cmd.exe) as administrator if possible
2. Run the command: ```winget install msys2.msys2```
3. Start ucrt64.exe from "C:\msys64"
4. Run the command: ```pacman -S --needed base-devel mingw-w64-ucrt-x86_64-toolchain```
5. Add ```C:\msys64\ucrt64\bin``` to system path
6. Restart the terminal and confirm that installation works with command: ```gcc --version```
<br><br>

### Using SDL2 with Go on Windows
There are few requirements to use SDL2 with Go. You can follow [veandco tutorial](https://github.com/veandco/go-sdl2?tab=readme-ov-file#requirements) how to get it working.  
Or you can follow these steps:
1. Install GCC (described in the list above)  
2. Download and extract SDL2 libraries  
   - Go to https://github.com/libsdl-org/SDL/releases
   - Download ```SDL2-devel-[version]-mingw.zip```, version should be the newest one
   - Extract the downloaded package (or use .zip explorer like 7-zip)
   - Go to ```x86_64-w64-mingw32``` folder in downloaded package
   - Copy contents of the folder into ```C:\msys64\ucrt64\```
3. Optional but highly recomended: repeat step 2 for [SDL_ttf](https://github.com/libsdl-org/SDL_ttf/releases) and [SDL_img](https://github.com/libsdl-org/SDL_image/releases)
4. When publishing an app with SDL2 libraries used you need to include runtime libraries (.dll files) with application .exe
   - To include these libraries go to [SDL](https://github.com/libsdl-org/SDL/releases), [SDL_ttf](https://github.com/libsdl-org/SDL_ttf/releases) and [SDL_img](https://github.com/libsdl-org/SDL_image/releases)
   - Download ```SDL2-[version]-win32-x64.zip```, ```SDL2_ttf-[version]-win32-x64.zip``` and ```SDL2_image-[version]-win32-x64.zip```
   - Extract .dll files from downloaded packages and place them in a project folder

Warning: First compilation after including SDL libraries may take long time!
<br><br>

### Other stuff
Great tutorial series on creating games in Go: [Games With Go](https://www.youtube.com/watch?v=9D4yH7e_ea8&list=PLDZujg-VgQlZUy1iCqBbe5faZLMkA3g2x)
