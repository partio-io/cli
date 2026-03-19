# VHS tape file — https://github.com/charmbracelet/vhs
#
# Records a demo GIF of partio's core workflow.
#
# Prerequisites:
#   - partio installed and in PATH
#   - Run from the cli/ directory (which has checkpoint data)
#
# Usage:
#   cd cli && vhs demo.tape

Output assets/demo.gif

Set FontSize 18
Set Width 1200
Set Height 650
Set Theme "Catppuccin Mocha"
Set Padding 20
Set TypingSpeed 50ms

# Show version
Type "partio version"
Enter
Sleep 1.5s

# Check status
Type "partio status"
Enter
Sleep 2s

# List checkpoints
Type "partio rewind --list"
Enter
Sleep 3s

# Resume a previous session (launches Claude Code)
Type "partio resume 0658682d9c50"
Enter
Sleep 7s

# Accept Claude Code permission prompt
Enter
Sleep 30s
