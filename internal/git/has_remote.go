package git

// HasRemote returns true if the repository has a remote named "origin".
func HasRemote() bool {
	_, err := execGit("remote", "get-url", "origin")
	return err == nil
}
