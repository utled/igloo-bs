package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"snafu/db"
)

func getHomeDir() (string, error) {
	homePath, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return homePath, nil
}

func Main() error {
	homePath, err := getHomeDir()
	if err != nil {
		return err
	}

	servicePath := filepath.Join(homePath, ".snafu")
	fmt.Println("servicePath:", servicePath)

	if info, err := os.Stat(servicePath); os.IsNotExist(err) {
		fmt.Println("servicePath does not exist, creating it: ")
		os.MkdirAll(servicePath, os.ModePerm)
		fmt.Println("initializing database")
		db.InitializeDB(servicePath)
		fmt.Println("setup complete")
	} else if err != nil {
		return fmt.Errorf("error checking if service path exist: %w", err)
	} else if !info.IsDir() {
		return fmt.Errorf("conflicting service path was found. path exist but is not a directory%v", info.Name())
	}

	return nil
}
