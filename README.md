# focus or not?

<a href='http://www.recurse.com' title='Made with love at the Recurse Center'><img src='https://cloud.githubusercontent.com/assets/2883345/11325206/336ea5f4-9150-11e5-9e90-d86ad31993d8.png' height='20px'/></a>

Hotkey-based toggle for Active Window Tracking (also known as "X-Mouse" or "focus follows mouse") on Windows. Runs a nice little program in the background that stays in your system tray and can even run on startup. Hotkey defaults to **Ctrl** + **Alt** + **X**.

Currently tested and working on Windows 11.

### config

You can change your hotkeys in the config file located at `%APPDATA%\FocusOrNot\config.json`. 

```json
{
  "hotkey": ["ctrl", "shift", "t"]
}
```

<video src="https://recurse.zulipchat.com/user_uploads/13/wCcDmzGWGGD0Q8jMJjihGdm0/Zed_lTDPxUvw5D.mp4" width="320" height="240" controls></video>

## what

On Windows, there's a feature officially called "Active Window Tracking" (also known as "X-Mouse" or "focus follows mouse"). Enabling it makes it so that placing your cursor automatically focuses the window and brings it to the front, unlike the default where you have to click to bring it to focus. This app lets you toggle that option with a hotkey, currently set to **Ctrl** + **Alt** + **X**.

## why

While I prefer AWT for the most part, there are occasions when I'm working on multiple screens where I'd rather not lose focus of a window just because I moved my cursor around. This is especially true for fullscreen applications (hey, VALORANT), where moving your cursor to the other monitor exits the fullscreen game, which is also partly because of Windows' weird behavior surrounding fullscreen applications.

## how

Instead of forcing you to navigate deep into the Windows Registry or Control Panel, the app interacts directly with the Windows kernel via the user32.dll dynamic link library.

The application uses the `SystemParametersInfoW` function with the `SPI_GETACTIVEWINDOWTRACKING` flag to query the operating system and find out if focus-follows-mouse is currently turned on or off.

When triggered, it calls the same function with `SPI_SETACTIVEWINDOWTRACKING`. It passes along special system flags `(SPIF_UPDATEINIFILE and SPIF_SENDCHANGE)` which tell Windows to save this preference permanently to your user profile and immediately broadcast the change to all open windows so the effect is instantaneous.

When you toggle the "Run at startup" option, the app interacts with the Windows Registry under `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`. It programmatically discovers its own executable file path and writes it to this registry key so Windows launches it automatically when you boot up your PC, or deletes the key cleanly if you uncheck the option.

## todo

- [X] Systray icons
- [X] Systray menu
- [ ] Option to change hotkey
- [ ] Executable icon
- [ ] Cross-platform??

## License

MIT.
