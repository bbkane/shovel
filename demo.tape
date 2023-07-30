# https://github.com/charmbracelet/vhs

Require shovel
Require bat

Output demo.gif

Set Shell zsh
Set PlaybackSpeed 0.5
Set TypingSpeed 100ms

Set FontSize 22
Set Height 1000
Set Width 1750
Set Padding 50

# https://github.com/charmbracelet/vhs/issues/39#issuecomment-1296334921
Hide
Sleep 5s
Show

Type "bat ~/.config/shovel.yaml"
Sleep 200ms
Enter
Sleep 8s

Type "shovel dig combine -q www.linkedin.com -r CNAME -s china -s usa -c 10"
Sleep 200ms
Enter
Sleep 8s

Sleep 3s