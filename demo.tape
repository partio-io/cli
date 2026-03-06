# VHS tape file — https://github.com/charmbracelet/vhs
#
# Usage:
#   vhs demo.tape
#
# Prerequisites:
#   - partio installed (make install)
#   - A git repo with at least one checkpoint
#   - VHS installed (brew install vhs)

Output assets/demo.gif

Set FontSize 16
Set Width 900
Set Height 500
Set Theme "Catppuccin Mocha"
Set Padding 20

# Setup: init a demo repo
Type "mkdir /tmp/partio-demo && cd /tmp/partio-demo && git init"
Enter
Sleep 1s

# Enable partio
Type "partio enable"
Enter
Sleep 2s

# Simulate a commit (user should have real checkpoint data)
Type "echo 'hello world' > main.go && git add . && git commit -m 'add main.go'"
Enter
Sleep 2s

# Check status
Type "partio status"
Enter
Sleep 2s

# List checkpoints
Type "partio rewind --list"
Enter
Sleep 3s

# Resume with --print
Type "partio resume $(partio rewind --list | head -1 | awk '{print $1}') --print"
Enter
Sleep 4s

# Done
Sleep 2s
