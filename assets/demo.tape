# Records assets/demo.gif. From the repo root: vhs assets/demo.tape
# (brew install vhs). The pacing and trimming are in demo-reel.sh. The
# desktop app's GIF, assets/desktop.gif, is made by desktop-demo.sh.

Output assets/demo.gif

Set Shell "bash"
Set FontSize 15
Set Width 1180
Set Height 760
Set Padding 28
Set TypingSpeed 70ms
Set Theme { "name": "Skyra", "background": "#080a0e", "foreground": "#c9ced6", "cursor": "#7ec4cc", "selection": "#1c2530", "black": "#080a0e", "red": "#e8a060", "green": "#8cc896", "yellow": "#e8c878", "blue": "#7ea4cc", "magenta": "#a898d8", "cyan": "#7ec4cc", "white": "#c9ced6", "brightBlack": "#5a6270", "brightRed": "#f0b478", "brightGreen": "#a4d8ac", "brightYellow": "#f0d890", "brightBlue": "#98bce0", "brightMagenta": "#bcaee8", "brightCyan": "#98d8e0", "brightWhite": "#eef1f5" }

Hide
Type "source assets/demo-reel.sh; clear"
Enter
Sleep 1s
Show

Sleep 600ms
Type "skyra demo"
Sleep 500ms
Enter
Sleep 25s
