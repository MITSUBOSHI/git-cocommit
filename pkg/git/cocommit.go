package git

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/MITSUBOSHI/cocommit/pkg/github"
)

const (
	coAuthoredByPrefix = "Co-Authored-By: "
)

// Cocommit executes git commit command with
// adding Co-Authored-By: to the commit message
func Cocommit(args []string) error {
	// Get Co-Authored-By: information
	coAuthors, err := getCoAuthors()
	if err != nil {
		return err
	}

	// Prepare arguments for git commit command
	commitArgs := []string{"commit"}
	messageIndex := -1
	hasMessageFlag := false

	// Find the position of -m flag
	for i, arg := range args {
		if arg == "-m" && i+1 < len(args) {
			messageIndex = i + 1
			hasMessageFlag = true
		}
		commitArgs = append(commitArgs, arg)
	}

	// If -m flag exists, modify the message
	if hasMessageFlag && messageIndex != -1 {
		// Add each coAuthors entry to the message
		message := args[messageIndex] + "\n\n"
		for _, coAuthor := range coAuthors {
			message += coAuthoredByPrefix + coAuthor + "\n"
		}
		commitArgs[messageIndex] = strings.TrimRight(message, "\n")

		// Execute git commit command
		cmd := exec.Command("git", commitArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		return cmd.Run()
	} else {
		// If no -m flag, implement editor flow
		return handleEditorCommit(args, coAuthors)
	}
}

// handleEditorCommit supports commit message editing using an editor
func handleEditorCommit(args []string, coAuthors []string) error {
	// Create a temporary commit message file
	tempFile, err := os.CreateTemp("", "COMMIT_EDITMSG")
	if err != nil {
		return fmt.Errorf("failed to create temporary commit message file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	// Get current branch name
	branchName := getCurrentGitBranch()

	// Write commit message template
	_, err = tempFile.WriteString(fmt.Sprintf("\n\n# Note: Co-Authored-By trailers will be automatically added to your commit message.\n\n# Please enter the commit message for your changes. Lines starting\n# with '#' will be ignored, and an empty message aborts the commit.\n#\n# On branch %s\n#\n", branchName))
	if err != nil {
		return fmt.Errorf("failed to write template to commit message file: %w", err)
	}
	tempFile.Close()

	// Open commit message file in the editor
	editor := getEditor()
	cmd := exec.Command(editor, tempFile.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}

	// Read the edited commit message
	messageContent, err := os.ReadFile(tempFile.Name())
	if err != nil {
		return fmt.Errorf("failed to read commit message: %w", err)
	}

	// Remove comment lines and trim whitespace
	message := cleanCommitMessage(string(messageContent))
	if message == "" {
		return errors.New("aborted commit due to empty commit message")
	}

	// Add Co-Authored-By
	message += "\n\n"
	for _, coAuthor := range coAuthors {
		message += coAuthoredByPrefix + coAuthor + "\n"
	}
	message = strings.TrimRight(message, "\n")

	// Update temporary file with modified message
	err = os.WriteFile(tempFile.Name(), []byte(message), 0644)
	if err != nil {
		return fmt.Errorf("failed to update commit message file: %w", err)
	}

	// Execute git commit command (read message from file)
	commitArgs := []string{"commit"}
	for _, arg := range args {
		// Do not use -m flag
		if arg != "-m" && !strings.HasPrefix(arg, "-m") {
			commitArgs = append(commitArgs, arg)
		}
	}
	commitArgs = append(commitArgs, "-F", tempFile.Name())

	cmd = exec.Command("git", commitArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// cleanCommitMessage removes comment lines from commit message and trims excess whitespace
func cleanCommitMessage(message string) string {
	var lines []string
	for _, line := range strings.Split(message, "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "#") {
			lines = append(lines, line)
		}
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// getEditor gets the editor to use
func getEditor() string {
	if editor := os.Getenv("GIT_EDITOR"); editor != "" {
		return editor
	}
	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor
	}
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}

	// Get editor from git config
	cmd := exec.Command("git", "config", "--get", "core.editor")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		return strings.TrimSpace(string(output))
	}

	// Default editor
	return "vi"
}

// readYesNo reads a y/n input
func readYesNo(prompt string) (bool, error) {
	fmt.Print(prompt + " (y/n): ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	input = strings.ToLower(strings.TrimSpace(input))
	return input == "y" || input == "yes", nil
}

// getCurrentGitUser gets the current Git user information (name and email address)
func getCurrentGitUser() (string, error) {
	return getCurrentGitUserFunc()
}

// Actual implementation
func getCurrentGitUserImpl() (string, error) {
	// Get user name
	nameCmd := exec.Command("git", "config", "--get", "user.name")
	name, err := nameCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git user name: %w", err)
	}

	// Get email address
	emailCmd := exec.Command("git", "config", "--get", "user.email")
	email, err := emailCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git user email: %w", err)
	}

	// Format
	userName := strings.TrimSpace(string(name))
	userEmail := strings.TrimSpace(string(email))

	return fmt.Sprintf("%s <%s>", userName, userEmail), nil
}

// Initialize function pointer
var getCurrentGitUserFunc = getCurrentGitUserImpl

// getGitAuthors gets unique author information from the Git commit history
// excluding the executor's own account
func getGitAuthors() ([]string, error) {
	// Get current user information
	currentUser, err := getCurrentGitUser()
	if err != nil {
		return nil, err
	}

	// Get author list with git command
	cmd := exec.Command("git", "log", "--format=%an <%ae>")
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to get git authors: %w", err)
	}

	// Split results by line, remove duplicates and current user
	lines := strings.Split(out.String(), "\n")
	uniqueAuthors := make(map[string]bool)
	var authors []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Add only if not empty, not duplicate, and not current user
		if line != "" && !uniqueAuthors[line] && line != currentUser {
			uniqueAuthors[line] = true
			authors = append(authors, line)
		}
	}

	return authors, nil
}

// isPecoAvailable checks if peco is available on the system
func isPecoAvailable() bool {
	_, err := exec.LookPath("peco")
	return err == nil
}

// selectWithPeco selects items using incremental search with peco
func selectWithPeco(items []string, prompt string) ([]string, error) {
	// Write choices to a temporary file
	tempFile, err := os.CreateTemp("", "cocommit-peco")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tempFile.Name())

	for _, item := range items {
		fmt.Fprintln(tempFile, item)
	}
	tempFile.Close()

	// Execute peco command
	cmd := exec.Command("peco", "--prompt", prompt)
	cmd.Stdin, err = os.Open(tempFile.Name())
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	// Get selection results
	selected := strings.Split(strings.TrimSpace(out.String()), "\n")
	var result []string
	for _, s := range selected {
		if s != "" {
			result = append(result, s)
		}
	}

	return result, nil
}

// selectFromList selects items in a standard way from a list
func selectFromList(items []string, prompt string) ([]string, error) {
	// Display item list
	fmt.Println(prompt)
	for i, item := range items {
		fmt.Printf("%d. %s\n", i+1, item)
	}

	// Accept selection input
	fmt.Print("Enter numbers (comma-separated) or 'all' for all items: ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	input = strings.TrimSpace(input)
	if strings.ToLower(input) == "all" {
		return items, nil
	}

	// Process comma-separated numbers
	selected := []string{}
	for _, numStr := range strings.Split(input, ",") {
		numStr = strings.TrimSpace(numStr)
		if numStr == "" {
			continue
		}

		var num int
		_, err := fmt.Sscanf(numStr, "%d", &num)
		if err != nil || num < 1 || num > len(items) {
			return nil, fmt.Errorf("invalid selection: %s", numStr)
		}

		selected = append(selected, items[num-1])
	}

	if len(selected) == 0 {
		return nil, errors.New("no valid selections made")
	}

	return selected, nil
}

// getCoAuthors gets Co-Authors information
// Gets GitHub usernames from GIT_COAUTHORS environment variable or standard input,
// and auto-completes email addresses using the GitHub API
func getCoAuthors() ([]string, error) {
	var usernames []string
	var result []string

	// Try to get from environment variable
	if coAuthors := os.Getenv("GIT_COAUTHORS"); coAuthors != "" {
		// Split comma-separated usernames into slice
		for _, username := range strings.Split(coAuthors, ",") {
			username = strings.TrimSpace(username)
			if username != "" {
				usernames = append(usernames, username)
			}
		}
	} else {
		// Select input method
		reader := bufio.NewReader(os.Stdin)

		fmt.Println("Select co-author input method:")
		fmt.Println("1. Manual input")
		fmt.Println("2. Select from Git history")
		fmt.Print("Enter your choice (1-2): ")

		// Read input using bufio.Reader
		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read input: %w", err)
		}

		input = strings.TrimSpace(input)
		choice := 1 // Default is manual input

		// Convert to number
		if input == "2" {
			choice = 2
		}

		if choice == 1 {
			// Get multiple users from standard input
			for {
				// Prompt for user input
				fmt.Print("Enter GitHub username: ")
				input, err := reader.ReadString('\n')
				if err != nil {
					return nil, err
				}

				// Validate input
				username := strings.TrimSpace(input)
				if username == "" {
					// Error if no one is specified
					if len(usernames) == 0 {
						return nil, errors.New("at least one GitHub username is required")
					}
					break
				}

				usernames = append(usernames, username)

				// Confirm adding additional users
				more, err := readYesNo("More co-author?")
				if err != nil {
					return nil, err
				}

				if !more {
					break
				}
			}
		} else {
			// Get Author information from Git history
			authors, err := getGitAuthors()
			if err != nil {
				return nil, err
			}

			if len(authors) == 0 {
				return nil, errors.New("no author information found in Git history")
			}

			// Select from Author information
			var selected []string
			if isPecoAvailable() {
				// Select with incremental search using peco
				selected, err = selectWithPeco(authors, "Select co-authors")
			} else {
				// Standard selection method
				selected, err = selectFromList(authors, "Available co-authors from Git history:")
			}

			if err != nil {
				return nil, err
			}

			// Use selected Author information
			return selected, nil
		}
	}

	// If no username is specified
	if len(usernames) == 0 {
		return nil, errors.New("at least one GitHub username is required")
	}

	// Get email address for each username and create Co-Authored-By format string
	for _, username := range usernames {
		email, err := github.GetUserEmail(username)
		if err != nil {
			return nil, fmt.Errorf("failed to get GitHub user information for '%s': %w", username, err)
		}

		// Create Co-Authored-By format string
		result = append(result, github.FormatCoAuthor(username, email))
	}

	return result, nil
}

// getCurrentGitBranch gets the current Git branch name
func getCurrentGitBranch() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		// Return default value if error occurs
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}
