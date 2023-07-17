# di-tui
A simple terminal UI player for [di.fm Premium](http://di.fm)

![App Screenshot](https://user-images.githubusercontent.com/3331648/81481515-bb668400-91fe-11ea-8a7c-39e1bb76c55d.png)

## Dependencies

### PulseAudio

Both linux and MacOS depend on pulseaudio to be running.

#### MacOS

By default, pulseaudio on MacOS runs as "root", which is not ideal. PulseAudio is best run by non-root users. By symbolically linking the pulseaudio plist file into your user's `~/Library/LaunchAgents/`, it runs as your user.

```
brew install pulseaudio
ln -s $(brew info pulseaudio | grep "/usr/local/Cellar" | awk '{print $1}')//homebrew.mxcl.pulseaudio.plist ~/Library/LaunchAgents
brew services start pulseaudio
```

#### Debian / Ubuntu

`apt install pulseaudio`

## Install

### Binary Releases

There are binary builds available in [releases](https://github.com/acaloiaro/di-tui/releases).

### With `go install`
`go install github.com/acaloiaro/di-tui@latest`

If `$GOPATH/bin` is not on your `$PATH` (modify accordingly for ZSH users `~/.zshrc`)
```
echo "export PATH=$PATH:$GOPATH/bin" >> ~/.bashrc
source ~/.bashrc
```

If your `$GOPATH` is not set, see https://github.com/golang/go/wiki/SettingGOPATH

## Authenticate

There are two authentication options

- Enter your username and password directly into `di-tui`
- If you're justifiably uncomfortable with entering your username/password into this application, copy your "Listen Key" from (https://www.di.fm/settings) and create the following file:

### ~/.config/di-tui/config.yml
```yml
token: <YOUR LISTEN KEY>
album_art: <BOOLEAN>
```

| key | description |
| --- | ----------- |
| token | Your di.fm authentication "Listen Key" found at https://www.di.fm/settings |
| album_art | Turn album art on or off |

## Configuration

### Themes

By default, `di-tui` respects your terminal's color scheme. However, there are four color settings that one can change by adding a `theme` to `config.yml`.

**Tomorrow-Night** inspired theme

```yml
theme:
  primary_color: "#81a2be"
  background_color: "#2a1f1a"
  primary_text_color: "#969896"
  secondary_text_color: "#81a2be"
```

## Run

`di-tui`

