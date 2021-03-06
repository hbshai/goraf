A program editing server for Radio AF written in Go.
![goraf ui](https://cloud.githubusercontent.com/assets/6011225/22562758/30bb5fdc-e97e-11e6-89ad-997cf2af113b.png)
usage
-----
assuming you have Go set up

 - `go get github.com/hbshai/goraf`
 - `cd $GOPATH/src/github.com/hbshai/goraf`
 - `go build`
 - `./goraf`
 - browse to localhost:8000

alternatively, you can get up-to-date builds from [gobuilder](https://gobuilder.me/github.com/hbshai/goraf).

we also tag and build release versions, found [here](https://github.com/hbshai/goraf/releases).

specifics
---------
This is a http server that serves a html site and allows users to edit a json file.

Only one user *should* be able to edit the json file at any given time. This is
accomplished with 1 session bound to an IP address + a session timeout.

The server will try to backup the program file before writing changes to it, but
if it, for some reason, fails the program will write changes and skip the backup.

Pressing delete will only remove the program for you. Click the save button to
update the server. This is for some additional safety, so that in the event of
incorrect deletion you only need to refresh your browser.

improvements
------------

 - validate input before submitting to server

thanks
------

 - sweetalert [link](https://github.com/t4t5/sweetalert)
