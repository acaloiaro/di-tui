# di-tui
A simple terminal UI player for [di.fm Premium](http://di.fm) 

![App Screenshot](https://user-images.githubusercontent.com/3331648/81481515-bb668400-91fe-11ea-8a7c-39e1bb76c55d.png)
# Caveat

This player is a somewhat crude proof-of-concept that was written over the course of ~8 hours and slightly improved upon since. There are not doubt bugs. Help me improve it if you find it useful. 

# Install

This app has been tested on Linux and Mac, but not Windows. However, it should also build on Windows. 

## Binary Releases

There are binary builds available in [releases](https://github.com/acaloiaro/di-tui/releases). 

## With `go get`
`go get -u github.com/acaloiaro/di-tui`

If `$GOPATH/bin` is not on your `$PATH` (modify accordingly for ZSH users `~/.zshrc`)
```
echo "export PATH=$PATH:$GOPATH/bin" >> ~/.bashrc
source ~/.bashrc
```

If your `$GOPATH` is not set, see https://github.com/golang/go/wiki/SettingGOPATH

# Authenticate

`di-tui --username "you@yourdomain.com" --password "your password"`

# Run

`di-tui`
