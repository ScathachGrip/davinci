# Developer & AI Agent Rules for DaVinci Bot

This file outlines critical implementation requirements and constraints for the DaVinci GitHub Bot to prevent breaking repository commit history, pull request links, and contributor attribution.

---

## 1. Pull Request Merge Strategy & Title Formatting

> [!IMPORTANT]
> **Use a tagged `switch mergeMethod` to format `CommitTitle` in `github.PullRequestOptions`** when performing a pull request merge.

- **`merge` method**: Set `opts.CommitTitle = pr.GetTitle()` to use the clean PR title and avoid GitHub's default `Merge pull request #...` prefix.
- **`squash` method**: Set `opts.CommitTitle = fmt.Sprintf("%s (#%d)", pr.GetTitle(), prNum)` to include the PR number link.
- **`rebase` method**: Leave `opts.CommitTitle` empty (`""`) to preserve individual commit titles.
- **Implementation**: Always use a tagged `switch mergeMethod` block to satisfy Go linter rules.

---

## 2. Co-authored-by Attribution & Bot User ID

> [!IMPORTANT]
> **Do not use the App ID as the Git user ID** in the `Co-authored-by` email.

- **Bot Email Format**: The email address for the bot co-author must match GitHub's official pattern:
  `{BOT_USER_ID}+{BOT_LOGIN}@users.noreply.github.com`
- **Retrieving the Correct User ID**:
  - The **App ID** (which is used to initialize/authenticate the bot client) is different from the **Bot's User ID** (which represents the database ID of the user account `slug[bot]`).
  - To get the correct **Bot User ID**, you must fetch the user profile from the GitHub API using:
    `client.Users.Get(ctx, botLogin)` (where `botLogin` is `{slug}[bot]`).
  - Caching or using the App ID instead of the Bot User ID causes GitHub to fail to map the email to the bot user, resulting in a grey, unlinked GitHub logo (Octocat) in the commit log instead of the bot's custom avatar.

---

## 3. Merge Method Auto-Detection

- **Detection Logic**: The bot's `detectLastMergeMethod` checks the repository history to decide whether to merge, squash, or rebase next. It parses the parent count and checks if the commit message contains `(#PR_NUMBER)`.
- If PR titles are merged without their native `(#PR_NUMBER)` suffix, this detection logic breaks (e.g. diagnosing squash merges as rebase merges), causing cascading bugs in subsequent PR merges. Maintain standard title formats to preserve detection accuracy.
