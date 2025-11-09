package paths

import (
    "os"
    "path/filepath"
)

var (
    BaseDir    = "."
    StorageDir = filepath.Join(BaseDir, "storage")
    UploadsDir = filepath.Join(StorageDir, "uploads")
    TextDir    = filepath.Join(StorageDir, "text")
    UsersFile  = filepath.Join(StorageDir, "users.json")
)

func EnsureDirs() error {
    for _, d := range []string{UploadsDir, TextDir} {
        if err := os.MkdirAll(d, 0755); err != nil {
            return err
        }
    }
    return nil
}