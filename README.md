Collection of smaller projects created while learning and exploring Go.
<br><br>

#### Easiest way (in my opinion) to install GCC (MSYS2 and MinGW) on Windows:
1. Open terminal (cmd.exe)
2. Run the command: ```winget install msys2.msys2```
3. Start ucrt64.exe from "C:\msys64"
4. Run the command: ```pacman -S --needed base-devel mingw-w64-ucrt-x86_64-toolchain```
5. Add "C:\msys64\ucrt64\bin" to path
6. Restart the terminal and confirm that installation works with command: ```gcc --version```
<br><br>

Great tutorial series on creating games in Go: [Games With Go](https://www.youtube.com/watch?v=9D4yH7e_ea8&list=PLDZujg-VgQlZUy1iCqBbe5faZLMkA3g2x).
