package sound

/*
#cgo LDFLAGS: -l winmm
#include <windows.h>
#include <stdio.h>

#pragma comment(lib, "Winmm.lib")

int MCIExecute(const char* cmd) {
	return mciSendString(cmd, "", 0, NULL);
}
*/
import "C"

import (
	"fmt"
)

func Execute(cmd string) int {
	code := C.MCIExecute(C.CString(cmd))
	return int(code)
}

type Player struct {
	Filename string
	Alias    string
}

func (player *Player) Load() int {
	code := Execute(fmt.Sprintf(`open "%s" alias %s`,
		player.Filename,
		player.Alias))
	return code
}

func (player *Player) Reload(filename string) int {
	player.Filename = filename
	return player.Load()
}

func (player *Player) Play() int {
	code := Execute(fmt.Sprintf("play %s from 0", player.Alias))
	return code
}

func (player *Player) Pause() int {
	code := Execute("pause " + player.Alias)
	return code
}

func (player *Player) Resume() int {
	code := Execute("resume " + player.Alias)
	return code
}

func (player *Player) Stop() int {
	code := Execute("stop " + player.Alias)
	return code
}

func (player *Player) JumpTo(time int) int {
	_ = Execute("stop " + player.Alias)
	code := Execute(fmt.Sprintf("play %s from %d", player.Alias, time))
	return code
}

func (player *Player) SetVolume(volume int) int {
	code := Execute(fmt.Sprintf("setaudio %s volume to %d", player.Alias, volume))
	return code
}

func (player *Player) Close() int {
	code := Execute("close " + player.Alias)
	return code
}
