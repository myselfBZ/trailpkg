package manifest

import (
	"encoding/json"
	"io"
	"os"
	"time"
)

type state struct {
	LastManifestUpdate time.Time       `json:"last_manifest_update"`
	Installed          []InstallRecord `json:"installed"`
}

type InstallRecord struct {
	PackageName string    `json:"name"`
	Version     string    `json:"version"`
	InstalledAt time.Time `json:"installed_at"`
}

func NewStateManager(filepath string) *StateManager {
	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {
		panic(err)
	}

	var parsedState state

	if err := json.NewDecoder(file).Decode(&parsedState); err != nil {
		panic(err)
	}

	return &StateManager{
		file:  file,
		state: parsedState,
	}
}

type StateManager struct {
	file  *os.File
	state state
}

func (i *StateManager) Append(record InstallRecord) {
	i.state.Installed = append(i.state.Installed, record)
}

func (i *StateManager) Remove(name string) {

	for idx := range len(i.state.Installed) {
		if i.state.Installed[idx].PackageName == name {
			i.state.Installed = append(i.state.Installed[:idx], i.state.Installed[idx+1:]...)
			break
		}
	}

}

func (i *StateManager) Find(name string) (InstallRecord, bool) {
	for idx := range len(i.state.Installed) {
		if i.state.Installed[idx].PackageName == name {
			return i.state.Installed[idx], true
		}
	}

	return InstallRecord{}, false
}

func (i *StateManager) UpdateManifestTime(newTime time.Time) {
	i.state.LastManifestUpdate = newTime
} 

func (i *StateManager) Save() error {
	i.file.Seek(0, io.SeekStart)
	i.file.Truncate(0)
	return json.NewEncoder(i.file).Encode(i.state)
}
