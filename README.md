# winpos

Windows positions store / restore command-line for Windows OS.

### Problem

Having two different monitors, I always have a chaos on my desktop 
after Windows go sleep / awake.

Need simple tool which will keep and eye on all 'normal' window 
positions and store it somewhere.

Then, after awake Windows, you can easy restore positions by short commands.

Looking for existing software, I can't find anything which makes 
my life more easy.


### Details

1. Logging. `winpos` is program without UI, it works hidden 
and don't show nothing to screen / console. Only way to see what happens
is log file. Log file created in running directory. So best to launch 
winpos from temporary dir, where log can be saved.

2. Windows data stored to file windows.txt near binary. It's not a modern, 
so please change it if you want so. PR is welcome.

3. `winpos` is Windows only software.


### Usage

`winpos s` or `winpos save` - saves position of all active windows.
`winpos r` or `winpos restore` - saves position of all active windows.


### TODO

[ ] Implement `watch` mode: program running in background, 
    periodically (on 5 seconds of user idle, for example), save all positions