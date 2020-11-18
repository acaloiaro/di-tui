# di-tui
A simple terminal UI player for [di.fm Premium](http://di.fm)

![App Screenshot](https://user-images.githubusercontent.com/3331648/81481515-bb668400-91fe-11ea-8a7c-39e1bb76c55d.png)

# Dependencies

## PulseAudio

Both linux and MacOS depend on pulseaudio to be running.

### MacOS

By default, pulseaudio on MacOS runs as "root", which is not ideal. PulseAudio is best run by non-root users. By symbolically linking the pulseaudio plist file into your user's `~/Library/LaunchAgents/`, it runs as your user.

```
brew install pulseaudio
ln -s $(brew info pulseaudio | grep "/usr/local/Cellar" | awk '{print $1}')//homebrew.mxcl.pulseaudio.plist ~/Library/LaunchAgents
brew services start pulseaudio
```

# Install

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
