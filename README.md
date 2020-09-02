# Diabotical master server POC

This emulates the Diabotical master server and allows to start the game while it's offline and even play multiplayer. But a bunch of additional stuff needs to be done to actually achieve this..

This was done on game version `0.20.297c` while official master servers were down.

## Building and running

* `./certs.sh`
* `env GOOS=linux GOARCH=amd64 go build main.go`
* `chmod +x main`
* `./main`

The master server talks TLS to the client, so make sure to create a cert and put it inside a `certs/` folder next to the master server binary.

## Rerouting traffic to the custom master server

There are two methods to do this.

### DNAT/SNAT on a Linux router

```
iptables -t nat -A  PREROUTING -d 188.122.71.80 -j DNAT --to-destination 192.168.0.1
iptables -t nat -A POSTROUTING -s 192.168.0.1 -j SNAT --to-source 188.122.71.80
```

Where `192.168.0.1` is the custom master servers IP.

When using this method, the game needs to be launched using the Epic Launcher, because it wants a Epic Games Service connection before it tries to contact the master server.

### Patching diabotical.exe

The official master server IP is hardcoded inside diabotical.exe, which can simply be patched with a hex editor. It's at offset `00839B68`.

The game now still needs to be launched using the Epic Launcher, which launches the game through anti-cheat. Although the anti-cheat doesn't care about the changed binary, I still wanted to try to disable it.

To change this, we can patch the diabotical.exe further to not require to be launched using Epic Launcher and Diabotical-Launcher.exe (which is the anti-cheat):

| Circumvent | Offset | Change from | To |
|--|--|--|--|
| Anti-Cheat | 000E5A9A | 74 (JZ) | 75 (JNZ) |
| Epic | 00145742 | 84 (JZ) | 85 (JNZ) |

This does most likely not _really_ circumvent the anti-cheat, but just disables the requirement to be launched using the launcher (which is all I'm trying to do here, don't cheat in multiplayer games).

Now we can launch the game just by opening the altered diabotical.exe.

## Usage and how it works

### Unlocking the client

The client handles the responses from the master server at offset `001429F0`.
The client asks the server on game launch for authorization by sending ASCII `!` over the open TCP socket. If the server answers with `4` and `#`, the client unlocks itself.

Note: Some of the commands the client sends to the master server can be found [here](https://github.com/ScheduleTracker/DiaboticalTracker/blob/master/ui/html/javascript/ui_data.js#L11), where the number represents an ASCII code.

At this point, the user is allowed to load and edit maps.

### Hosting a server

Loading a map with `/map <mapname>` starts a LAN server, so other people are able to connect to it. But sadly, the `/connect` command was removed from Diabotical to prevent simply connecting to it.
There's still a way to connect (and I forgot how I did it..), but it's not really useful anyway, because starting a server with `/map <mapname>`, the game doesn't set the game parameters (round limit, etc.) correctly.

The other way to start a server is to ask the master server for a LAN server instance, which in turn tells the client to launch a server. To do this (when connected to the custom master server), type `start <game_mode> <map>` into the main menu chat. This then launches a local server on port `32123` with the game parameters correctly set.

You could now setup port forwarding on your router to make it publicly accessible and play online.

### Joining a server

The master server saves the game servers IP, game_mode and map in memory. So at this point, a client can connect by typing `join` into the main menu chat.

Note that the custom master server can only handle one game server at a time (`start`ing another game server overwrites the parameters of the first one).

### Changing Game mode

All good, right? No.

Launching a LAN server using this method DOES correctly set game parameters but does NOT initialize the game mode correctly. This seems to be a bug or an incomplete game server implementation in the client.
The game can't find the game mode definition and therefore loads the definition of game mode `default`.

To work around this, we need to change the `default` [game mode definition](https://github.com/ScheduleTracker/DiaboticalTracker/blob/master/scripts/base.games#L3) to include all changes from the mode you want to play. All other modes seem to inherit their settings from the `default` definition. So keep the default definition as is and only add or change the lines where the desired game mode differs.

To actually change the file that contains the game mode definitions (`base.games`), you need to unpack the `packs/scripts.dbp` file, change `base.games`, then repack it. This can be done with [dbt-pack](https://github.com/marconett/dbt-packer).

Note that this only needs to be done on the game server.