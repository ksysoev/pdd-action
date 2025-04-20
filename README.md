# pdd-action
Github Action to add Puzzle Driven Development into your Github Repository 

Puzzle Driven Development (PDD) is a software development methodology that focuses on breaking down complex problems into smaller, manageable puzzles.
It encourages collaboration, creativity, and iterative problem-solving to deliver high-quality software solutions. 
PDD emphasizes the importance of understanding the problem domain and leveraging the collective intelligence of the development team to find innovative solutions.

How process works:
1. Developer start working on github issue
2. In many cases, issue doesn't cover all underlying complexity
3. Along Development process, developer creates TODO comments in the codebase, to highlight the additional work that needs to be done 
2. This tool will parse PR and find all TODO comments
3. Then tool will create new issues in the repository, based on the TODO comments
4. and update comments in the PR with ids of the new issues

Comments format before issue creation:
```
// TODO: {issue_title}
// Labels: {comma_separated_labels} (optional)
// {issue_description}
// {issue_description_continue}
```

Comments format after issue creation:
```
// TODO: {issue_title}
// Issue: {issue_url}
// Labels: {comma_separated_labels} (optional)
// {issue_description}
// {issue_description_continue}
```

This tool should support comments format for as many languages as possible, including:
GoLang, Java, Python, JavaScript, TypeScript, C#, C++, C, Ruby, Swift, Kotlin, Rust, PHP, HTML, CSS, Shell Script, Bash Script, PowerShell Script, SQL, R, Perl, Haskell, Scala, Groovy, Lua, Elixir, Erlang, F#, Objective-C

Github Action configuration should accept the following parameters:
- Github token to create issues in the repository
- Branch name to create issues in the repository, default is main. While PR is not yet merged issues should not be yet created.
- Github issue title prefix, default is none