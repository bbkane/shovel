# https://github.com/charmbracelet/vhs

Require example-go-cli

Output demo.gif

Set Shell zsh
Set PlaybackSpeed 0.5
Set TypingSpeed 75ms

Set FontSize 22
Set Height 650
Set Width 1500
Set Padding 50

Hide
Type 'setopt interactive_comments'
Enter
Type 'clear'
Enter
Show

Type "shovel dig combine --qname example.com --qname www.example.com \"
Enter
Type "--rtype A --nameserver 1.1.1.1:53 --count 10"
Enter
Sleep 1s

Sleep 3s
