# git-cocommit

A Git subcommand that automatically adds `Co-Authored-By:` to commit messages.

## Features

- The `git cocommit` command automatically adds `Co-Authored-By:` to your commit messages
- You only need to specify GitHub usernames through the `GIT_COAUTHORS` environment variable or standard input
- **Multiple users support**: Comma-separated in environment variables, or interactively add multiple users from standard input
- **Selection from Git History**: You can select author information from the repository's commit history
- **peco integration**: If peco is installed, you can use incremental search for selection
- Uses GitHub API to automatically retrieve email addresses from usernames
- Supports commit message specification with the `-m` flag, or editing commit messages using an editor

## Installation

### Prerequisites

- Git
- Go 1.23.0 or higher

### Optional Dependencies

- [peco](https://github.com/peco/peco) - If you want to use incremental search functionality

### Installation Steps

```bash
# Clone the repository
git clone https://github.com/MITSUBOSHI/cocommit.git
cd cocommit

# Install dependencies
go mod download

# Run the installation script
./install.sh
```

The installation script (`install.sh`) builds the program and places it as `bin/git-cocommit`.

### Setting up as a Git subcommand

Git can run a program as `git <subcommand>` if there is a program named `git-<subcommand>` in the PATH. Use one of the following methods to ensure git-cocommit is in your PATH:

#### Method 1: Create a symbolic link (recommended)

```bash
# Create a symbolic link (per user)
ln -sf "$(pwd)/bin/git-cocommit" ~/bin/git-cocommit

# Or, to make it available system-wide (requires admin privileges)
sudo ln -sf "$(pwd)/bin/git-cocommit" /usr/local/bin/git-cocommit
```

Note: Make sure `~/bin` is included in your PATH. If not, add the following to your shell configuration file (`.bashrc`, `.zshrc`, etc.):

```bash
export PATH="$HOME/bin:$PATH"
```

#### Method 2: Copy directly

```bash
# For per-user installation
mkdir -p ~/bin
cp bin/git-cocommit ~/bin/

# Or, for system-wide installation (requires admin privileges)
sudo cp bin/git-cocommit /usr/local/bin/
```

### Verifying the Installation

To verify that the installation was successful:

```bash
# Check if the command is available
which git-cocommit

# Display version information or help (if available)
git cocommit --help
```

## Usage

### Specification via Environment Variables

```bash
# Setting the environment variable - for a single GitHub username
export GIT_COAUTHORS="username"

# For multiple users (comma-separated)
export GIT_COAUTHORS="username1, username2, username3"

# Use instead of normal git commit (with -m flag)
git cocommit -m "Commit message"

# When using an editor (without -m flag)
git cocommit
```

### Interactive Input

If the environment variable is not set, you can choose the input method:

```
Select co-author input method:
1. Manual input
2. Select from Git history
Enter your choice (1-2): 
```

#### 1. Manual Input

How to manually input GitHub usernames:

```
Enter GitHub username: username1
More co-author? (y/n): y
Enter GitHub username: username2
More co-author? (y/n): y
Enter GitHub username: username3
More co-author? (y/n): n
```

#### 2. Select from Git History

How to select author information from the repository's commit history:

- If peco is installed: You can select using the incremental search feature.
  
  ![peco selection image](https://github.com/peco/peco/raw/master/doc/peco.gif)

- If peco is not installed: You can select using a number-based selection method.

  ```
  Available co-authors from Git history:
  1. User Name1 <user1@example.com>
  2. User Name2 <user2@example.com>
  3. User Name3 <user3@example.com>
  Enter numbers (comma-separated) or 'all' for all items: 1,3
  ```

### Generated Commit Message

This will create a commit message like:

```
Commit message

Co-Authored-By: username1 <1234567+username1@users.noreply.github.com>
Co-Authored-By: username2 <7654321+username2@users.noreply.github.com>
Co-Authored-By: username3 <9876543+username3@users.noreply.github.com>
```

### About GitHub API Usage

- When you only specify the username, GitHub API is used to retrieve user information
- If you set the `GITHUB_TOKEN` environment variable, authenticated API calls will be made, relaxing rate limits
- If no token is set, unauthenticated API calls will be made, but please be aware of rate limits
- If the user's public email address is not set, GitHub's no-reply email address (`ID+USERNAME@users.noreply.github.com` format) will be used
  - This email format complies with GitHub's official [privacy-protected email address](https://docs.github.com/en/account-and-profile/setting-up-and-managing-your-personal-account-on-github/managing-email-preferences/setting-your-commit-email-address) format

### Editor Configuration

By default, the editor is selected in the following order:

1. `GIT_EDITOR` environment variable
2. `VISUAL` environment variable
3. `EDITOR` environment variable
4. git's `core.editor` configuration
5. Default to `vi`

## Notes

- When editing a commit message in an editor, comment lines (lines starting with `#`) are ignored
- If you cancel the editor or enter an empty message, the commit will be aborted
- Go 1.23.0 or higher is required
- To use peco functionality, you need to install peco separately

## License

MIT 