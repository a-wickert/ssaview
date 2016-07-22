#ssaview
-------
TODO: provide new heroku url

ssaview is a small utlity that renders SSA code alongside input Go code

Runs via HTTP on :8080

License: ISC

```sh
  $ go get github.com/a-wickert/ssaview
  $ go install github.com/a-wickert/ssaview
  $ Optional: which ssaview should show you: $GOPATH/bin/ssaview
  $ ssaview &
  open localhost:8080
```

Screenshot:
![Example screenshot](https://github.com/tmc/ssaview/raw/master/.screenshot.png)

## TODO

* Logging not via fmt
* Fix that the stuff is shown twice
* Show error messages
* Show value of checkbox after rendering
