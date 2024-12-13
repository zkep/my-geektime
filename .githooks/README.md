### [golangci-lint](https://golangci-lint.run/usage/quick-start/)
```shell
git config core.hooksPath .githooks
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### error prompt
```shell
hint: The '.git/hooks/pre-commit' hook was ignored because it's not set as executable.
hint: You can disable this warning with `git config advice.ignoredHook false`.
```
```shell
chmod 777 .githooks/pre-commit
chmod 777 .githooks/pre-push
```