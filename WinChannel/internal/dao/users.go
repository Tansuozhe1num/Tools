package dao

import (
    "encoding/json"
    "os"
    "winchannel/internal/model"
    "winchannel/internal/paths"
)

var Users = &model.UserStore{Users: map[string]string{}}

func LoadUsers() {
    Users.Mu.Lock()
    defer Users.Mu.Unlock()
    if _, err := os.Stat(paths.UsersFile); err != nil {
        _ = os.MkdirAll(paths.StorageDir, 0755)
        f, _ := os.Create(paths.UsersFile)
        if f != nil {
            _ = json.NewEncoder(f).Encode(map[string]string{})
            f.Close()
        }
        Users.Users = map[string]string{}
        return
    }
    f, err := os.Open(paths.UsersFile)
    if err != nil {
        Users.Users = map[string]string{}
        return
    }
    defer f.Close()
    m := map[string]string{}
    if err := json.NewDecoder(f).Decode(&m); err != nil {
        Users.Users = map[string]string{}
        return
    }
    Users.Users = m
}

func SaveUsers() error {
    Users.Mu.Lock()
    defer Users.Mu.Unlock()
    tmp := Users.Users
    if err := os.MkdirAll(paths.StorageDir, 0755); err != nil {
        return err
    }
    f, err := os.Create(paths.UsersFile)
    if err != nil {
        return err
    }
    defer f.Close()
    return json.NewEncoder(f).Encode(tmp)
}