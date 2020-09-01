# Diabotical Masterserver POC

This emulates the Diabotical masterserver and allows to start the game while it's offline and even play multiplayer. But a bunch of additional stuff needs to be done to actually achieve this..

This was done on game version `0.20.297c` while official masterservers were down.

## Building and running

* `./certs.sh`
* `env GOOS=linux GOARCH=amd64 go build main.go`
* `chmod +x main`
* `./main`

The masterserver talks TLS to the client, so make sure to create one and put it inside a `certs/` folder next to the masterserver binary.

## Rerouting traffic to the custom Masterserver

There's two methods to do this.

### DNAT/SNAT on a Linux router

```
iptables -t nat -A  PREROUTING -d 188.122.71.80 -j DNAT --to-destination 192.168.0.1
iptables -t nat -A POSTROUTING -s 192.168.0.1 -j SNAT --to-source 188.122.71.80
```

Where `192.168.0.1` is the custom masterserver IP.

When using this method, the game needs to be launched using the Epic Launcher, because the it wants a Epic Games Service connection before it tries to contact the masterserver.

### Patching diabotical.exe

The official masterserver IP is hardcoded inside diabotical.exe, which can simply be patched. It's at offset `00839B68`.

The game still needs to be launched using the Epic Launcher, which then also runs through anti-cheat. Altough the anti-cheat doesn't care about the changed binary (it still just launches), it's not ideal.

To change this, we can patch the diabotical.exe further to not require to be launched using Epic Launcher and Diabotical-Launcher.exe (which is the anti-cheat):

| Circumvent | Offset | Change from | To |
|--|--|--|--|
| Anti-Cheat | 000E5A9A | 74 (JZ) | 75 (JNZ) |
| Epic | 00145742 | 84 (JZ) | 85 (JNZ) |

This does most likely not _really_ circumvent the anti-cheat, but just disables the requirement to be launched using the launcher (which is all I'm trying to do here, don't cheat in multiplayer games).

Now we can launch the game just from starting the diabotical.exe.

## Usage and how it works

### Unlocking the client

The client handles the responses from the masterserver at offset `001429F0`. The client asks the server on game launch for auth with `!`. If the server answers with `4` and `#`, the client unlocks itself.

Note: Some of the commands the client sends to the master server can be found [here](https://github.com/ScheduleTracker/DiaboticalTracker/blob/master/ui/html/javascript/ui_data.js#L11), where the number represents an ASCII code.

At this point, the user is allowed to load and edit maps.

### Hosting a server

Loading a map with `/map <mapname>` also starts a LAN server, so other people are able to connect to the server. Now sadly, the `/connect` command was removed from Diabotical to prevent things like this.
There's a way to connect anyways (and I simply forgot how I did it..), but it's not really useful anyways, because starting a server with `/map <mapname>`, the game doesn't set the game parameters (round limit, etc.) correctly.

The other way to start a server is ask the masterserver for a LAN server instance, which then tells the client exactly what to do. To to this (when connected to the custom masterserver), type `start <game_mode> <map>` into the main menu chat to launch a LAN server. This then launches a local server on port `32123` with the game parameters correctly set.

You could now setup port forwarding on your router to make it publicly accessible and play online.

### Joining a server

The masterserver saves the gameserver's IP, game_mode and map in memory. So at this point, Another client can connect by simply typing `join` into the main menu chat.

Note that the custom masterserver can only handle one gameserver at a time (`start`ing another gameserver, overwrites the parameters of the first one).

### Changing Game mode

All good, right? No.

Launching a LAN server using this method DOES correctly set game parameters but does NOT initialize the game mode correctly. This seems to be a bug or an incomplete server implementation in the client.
The game can't find the game mode definition and therefore loads the definition of game mode `default`.

To fix this, we need to change the `default` [game mode definition](https://github.com/ScheduleTracker/DiaboticalTracker/blob/master/scripts/base.games#L3) to include all changes from another mode. All other modes seem to inherit the settuings from the `default` definition. So keep the default definition as is, and only add or change the lines where the desired game mode differs.

To actually change the file that contains the game mode definitions (`base.games`), you need to unpack the `packs/scripts.dbp` file, change `base.games`, then repack it. This can be done with [dbt-pack](https://github.com/marconett/dbt-packer).

Note that this only needs to be done on the server.