package github

// EditIssue replaces the title and the body of issue number in repo.
// GitHub keeps the old text in the issue's edit history.
func (c Client) EditIssue(repo string, number int, title, body string) error {
	return c.patchIssue(repo, number, map[string]string{"title": title, "body": body}, "edit")
}
