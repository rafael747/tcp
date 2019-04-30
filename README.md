# tcp

## A simple tool for sending files over tcp

The goal is to send a file to another computer in the same network, as fast as possible

### How to install

    go get github.com/rafael747/tcp
    go install github.com/rafael747/tcp

> Make sure you have $GOPATH defined

  You can also use the pre-build package


### How to use

 - to receive a file in the working directory

       tcp

 - to send a file

       tcp file host


> the host can be a IP address or a name, check your DNS configuration


## BUGS and TODO

 - The hosts must be acessible directly, no firewalls or NATs
 - Currently, using only port 2000/tcp
 - Optional encryptation

