# di-tui
A simple terminal UI player for [di.fm Premium](http://di.fm)

![App Screenshot](https://user-images.githubusercontent.com/3331648/81481515-bb668400-91fe-11ea-8a7c-39e1bb76c55d.png)

# Install

## Dependencies 

## PulseAudio

Both linux and MacOS depend on pulseaudio to be running.

### MacOS

By default, pulseaudio on MacOS runs as "root", which is not ideal. PulseAudio is best run by non-root users. By symbolically linking the pulseaudio plist file into your user's `~/Library/LaunchAgents/`, it runs as your user.

```
brew install pulseaudio
ln -s $(brew info pulseaudio | grep "/usr/local/Cellar" | awk '{print $1}')//homebrew.mxcl.pulseaudio.plist ~/Library/LaunchAgents
brew services start pulseaudio
```

### Linux 

`apt install pulseaudio`

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

There are two authentication options 

- Enter your username and password directly into `di-tui`
- If you're justifiably uncomfortable with entering your username/password into this application, copy your "Listen Key" from (https://www.di.fm/settings) and create the following file:

## ~/.config/di-tui/config.yml 
```yml
token: <YOUR LISTEN KEY>
```


# Run

`di-tui`
