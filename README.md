#ssaview
-------

ssaview is a tool which renders a Go code into the SSA representation.
A online version is accesible via [heroku](https://powerful-earth-92559.herokuapp.com).

It is possible to add some additional information like the type of each instruction and the idoms of each basic block to the SSA representation via a checkbox.
An other possibility is that the build mode of the SSA can be changed from the standard mode to the SanityCheckFunctions mode.

The application starts on the port of the environment variable PORT.
If the variable is not set, the application start on port 8080.
It is possible to change the port by setting the environment variable e.g:
'$ export PORT=8080 '

License: ISC

```sh
  $ go get github.com/a-wickert/ssaview
  $ go install github.com/a-wickert/ssaview
  $ Optional: which ssaview should show you: $GOPATH/bin/ssaview
  $ ssaview &
  open localhost:8080
```

Screenshot:
![Example screenshot](https://github.com/akwick/ssaview/raw/master/.preview.png)

## TODO

- [x] Logging not via fmt
- [ ] Fix that the stuff is shown twice
- [ ] Show error messages
- [ ] Show value of checkbox after rendering
- [x] Provide heroku link
