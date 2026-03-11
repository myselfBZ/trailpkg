package manifest

import (
	"encoding/json"
	"os"
	"io"
	"time"
)


type state struct {
	// Suggestion: Could use a nice map[InstallRecord]bool as set here. 
	// Fine for now
	Installed []InstallRecord `json:"installed"`
}

type InstallRecord struct {
	PackageName string `json:"name"`
	Version 	string `json:"version"`
	InstalledAt time.Time `json:"installed_at"`
}


func NewInstallRecordManager(filepath string) *InstallRecordManager {
	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {
		panic(err)
	}

	var parsedState state

	if err := json.NewDecoder(file).Decode(&parsedState); err != nil {
		panic(err)
	}

	return &InstallRecordManager{
		file: file,
		state: parsedState,
	}
}

type InstallRecordManager struct {
	file *os.File
	state state
}

func (i *InstallRecordManager) Append(record InstallRecord) {
	i.state.Installed = append(i.state.Installed, record)
}

func (i *InstallRecordManager) Remove(name string) {

	for idx := range len(i.state.Installed) {
		if i.state.Installed[idx].PackageName == name {
			i.state.Installed = append(i.state.Installed[:idx], i.state.Installed[idx+1:]...)
			break
		}
	}

}

func (i *InstallRecordManager) Find(name string) (InstallRecord, bool) {
	for idx := range len(i.state.Installed) {
		if i.state.Installed[idx].PackageName == name {
			return i.state.Installed[idx], true
		}
	}

	return InstallRecord{}, false
}

func (i *InstallRecordManager) Save() error {
    i.file.Seek(0, io.SeekStart)
    i.file.Truncate(0)
    return json.NewEncoder(i.file).Encode(i.state)
}


