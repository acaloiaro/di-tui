# di-tui
A simple termainl UI player for [di.fm](http://di.fm)
---
![App Screenshot](https://user-images.githubusercontent.com/3331648/75633473-08145a00-5bd3-11ea-9c6c-0519abd8730b.png)
# Caveat

Currently this is very crude proof of concept that was written over the course of ~8 hours. There *are* bugs. Help me improve it if you find it useful. 

# Install

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
