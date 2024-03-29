# Toonily-dl
Download comics from [Toonily.com](https://toonily.com/) website

## Features
- [X] Download all available images of comics
- [X] Download comics in dedicated directory
- [X] Download chapter in dedicated directory
  * [ ] Download in asceding order
  * [X] Download picked chapters
- [X] Continue downloading, or update with new chapters/files
- [X] Download comic's cover image

## How to use
This application offers users the ability to run through either a simple bash script or a golang binary/executable, catering to different user preferences and system capabilities.based on their technical preferences and system requirements.
### Bash
```bash
>> ./toonily-dl.sh https://toonily.com/webtoon/amazing-manga/
Download: Amazing Manga
Working on chapter-9
Downloaded 12 file(s)
Working on chapter-8
Downloaded 11 file(s)
Working on chapter-7
...
```
### Golang
```cmd
C:\> toonily-dl.exe https://toonily.com/webtoon/amazing-manga/
Download: Amazing Manga
Working on chapter-9
Downloaded 12 file(s)
Working on chapter-8
Downloaded 11 file(s)
Working on chapter-7
...
```

```bash
SYNOPSIS
		toonily-dl [flag]... <URL>
	FLAGS
  
		-d
            Save description
		-c
		    Indicate the chapter's list to download.
		    Example:
			    toonily-dl -c 3         :: download only chapter 3
			    toonily-dl -c 13:       :: download chapters starting with 13 until the end
			    toonily-dl -c 213:321   :: download chapters starting with 213 to 321
			    toonily-dl -c :3210     :: download chapters up to 3210
```



### Result
![vokoscreenNG-2024-01-14_14-19-56 801](https://github.com/v613/ToonilyDownloader/assets/15879258/a956ac9c-b540-44f8-b8c6-326b52e8f5f2)

### Setup a better experience
To have a better experience, it is recommended to save, move, or copy it to the "/usr/local/bin" directory. This directory is typically included in the system's PATH, which allows you to execute the script from any location in your terminal.
In Windows platforms, you need to throw the executable into a directory that is already in the PATH, or add a new one.

Here is how you can do it:

1. Open a terminal.
2. Assuming you have the "toonily-dl.sh" script already downloaded or available on your system, navigate to the directory where the script is located.
3. Once you are in the directory with the script, you can use the following command to move it to the "/usr/local/bin" directory:
   ```bash
   sudo mv toonily-dl.sh /usr/local/bin/toonily
   ```
   This command will prompt you for your password, as it requires administrator privileges to move a file to the "/usr/local/bin" directory.
   > Make sure you have the necessary execution permissions for the script by running:
   >
   >```bash
   >sudo chmod 755 /usr/local/bin/toonily-dl
   >```
   >
   > This will make the script executable.

5. After providing your password, the script will be moved to the "/usr/local/bin" directory.

With the script in this directory, you can now execute it from any location in your terminal by simply typing "toonily-dl". 
