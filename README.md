# MIC-1 Emulator
A simple [Mic-1 Emulator](https://en.wikipedia.org/wiki/MIC-1) written in Go with a terminal UI.

## Features

* Terminal UI
* Memory inspector
* Register inspector
* Microcode inspector
* Microcode breakpoints

## Usage

The MIC-1 Emulator supports loading both binary and binary string microcode and memory files.

* -mc file
  * Loads the given binary microcode file
* -mcs file
  * Loads the given binary string microcode file
* -m file
  * Loads the given binary memory file
* -ms file 
  * Loads the given binary string memory file

## Screenshots

![Screenshot](img/main.png?raw=true)

## Key Bindings
### Global
Key Combination | Description
---|---
<kbd>q</kbd> | Quit
<kbd>CTRL + c</kbd> | Quit
<kbd>c</kbd> | Cycle frame focus forward direction
<kbd>SHIFT +  c</kbd> | Cycle frame focus reverse direction
<kbd>s</kbd> | Steps the MIC-1 emulator forward one complete cycle
<kbd>r</kbd> | Runs the MIC-1 emulator until a HALT is requested or a break point is hit
<kbd>h</kbd> | Halts the MIC-1 emulator
<kbd>l</kbd> | Resets the MIC-1 emulator. Stops execution, zeros memory and microcode, and reloads microcode and memory 

### Symbols Frame

Key Combination | Description
---|---
<kbd>j</kbd> | Scrolls down one symbol
<kbd>k</kbd> | Scrolls up one symbol
<kbd>ENTER</kbd> | Moves the memory frame to the symbol's location in memory
<kbd>g</kbd> | Moves the memory frame to the symbol's location in memory
<kbd>m</kbd> | Toggles the display mode between hexadecimal and decimal 

### Memory Frame

Key Combination | Description
---|---
<kbd>j</kbd> | Scrolls down by eight words
<kbd>k</kbd> | Scrolls up by eight words
<kbd>m</kbd> | Toggles the display mode between hexadecimal and decimal 

### Microcode Frame

Key Combination | Description
---|---
<kbd>j</kbd> | Scrolls down by one instruction
<kbd>k</kbd> | Scrolls up by one instruction
<kbd>b</kbd> | Toggles breakpoint on that instruction

## Todo
* Better scrolling
* Better colors
