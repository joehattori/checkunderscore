# checkunderscore
This static analysis tool checks if the returned value of each function is always ignored, or handled with `_`.
You can consider refactoring such functions.

## Example
```go
func getUsername() (string, error) {
  ...
}

var username, _ = getUsername()
...
func someOtherFunc() {
  user, _ := getUsername()
  ...
}
```
In this example, the function `getUsername()` may cause `error`, but the second return value `error` is always ignored with `_`.
`checkunderscore` responds with a warning that says `getUsername(): 2nd returned value is always ignored`.
