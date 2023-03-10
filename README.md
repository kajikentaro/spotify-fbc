# Spotify file-based Client

[(JAPANESE 日本語)](./README_jp.md)

## Features

**(1)** Download Spotify playlist information locally.

- Songs are represented as text files.
- Playlists are represented as folders.

<div style="text-align:center;">
↓
</div>

**(2)** Backup, move songs, add songs, etc.  
 (can be done in Explorer, etc.)

<div style="text-align:center;">
↓
</div>

**(3)** Transfer folder, text file status to Spotify.

<div style="text-align:center;">
Example of a text file of a song
</div>

```txt
id 4B0JvthVoAAuygILe3n4Bs
name What Do You Mean?
artist Justin Bieber
album Purpose (Deluxe)
seconds 205680
isrc USUM71511919
file_name What Do You Mean .txt
```

## How to use

### (1) Download

You can download Windows and Linux versions from here.  
https://github.com/kajikentaro/spotify-fbc/releases/tag/0.1

After downloading, it is recommended to rename it to spotify-fbc.  
(spotify-fbc.exe on Windows)

Run the following command in the file where spotify-fbc exists and make sure that help is displayed.

```
$ spotify-fbc help
Spotify file-based client:
Edit your playlists by moving directories and file locations

Usage:
  spotify-fbc [command]

Available Commands:
  clean       Clean up unused playlist entity txt
  compare     Compare local playlists with your spotify account and print the difference
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  login       Perform login process
  logout      Logout from your spotify account excluding API keys
  pull        Download playlists that your spotify account has. All of your existing local playlists will be overwritten
  push        Synchronize your local files and directories with your spotify account
  reset       Delete user-specific data such as OAuth token and Client ID excluding music txt
  version     Print the version number of spotify-fbc

Flags:
  -h, --help   help for spotify-fbc
```

### (2) Preparation (API key creation)

1. Go to https://developer.spotify.com/dashboard/applications and click the `CREATE AN APP` button.
2. Decide `App name` and `App description` as you like.
3. Check `Client ID`.
4. Click `SHOW CLIENT SECRET` button and check the `Client Secret`.
5. Click `EDIT SETTINGS` button, enter `http://localhost:8080/callback` in `Redirect URIs` and click `ADD`.  
   Then click `SAVE` to save the file.

### (3) Login

Execute the following command.

```
$ spotify-fbc-0.1-linux-amd64 login
```

Follow the output on the screen.  
If you see `token cache was saved to ~~~`, you have succeeded.

### (4) Download Spotify song information

Execute the following command.

```
$ spotify-fbc pull
```

The directory `spotify-fbc` will be created and the songs and playlists will be stored in it.

### (5) Compare difference information

Once you have made the necessary edits to the `spotify-fbc` directory, check the differences before `push`!

```
$ spotify-fbc compare
+ new playlist
  + What Do You Mean?
- deleted playlist
```

You will see that the `new playlist` playlist and the song `What Do You Mean?` are added, and the `deleted playlist` is removed,

### (6) Upload Spotify song information

Execute the following command.

```txt
$ spotify-fbc push
```

Your existing Spotify playlists will be completely replaced in the `spotify-fbc` directory.
Run carefully!

### (7) Searching for songs

A song info text file consists of several properties.

For example, if you create a new file like the following and do `spotify-fbc push`,

```text
name What Do You Mean?
artist Justin Bieber
```

The search will be performed with the song title as `What Do You Mean?` and the artist name as `Justin Bieber`.

Other properties such as `id`, `name`, `artist`, `album`, and `isrc` are also supported.  
(Other properties are for system administration.)
