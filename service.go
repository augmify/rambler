package main

import (
	"fmt"
	"github.com/elwinar/rambler/driver"
	_ "github.com/elwinar/rambler/driver/mysql"
	_ "github.com/elwinar/rambler/driver/postgresql"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Service is the struct that gather operations to manipulate migrations table
// and migrations on the filesystem.
type Service struct {
	driver.Conn
	env Environment
}

// NewService initialize a new service with the given informations
func NewService(env Environment) (*Service, error) {
	if _, err := os.Stat(env.Directory); err != nil {
		return nil, fmt.Errorf(`directory %s unavailable: %s`, env.Directory, err.Error())
	}

	conn, err := driver.Get(env.Driver, env.DSN(), env.Database)
	if err != nil {
		return nil, fmt.Errorf(`unable to initialize driver: %s`, err.Error())
	}

	return &Service{
		Conn: conn,
		env:  env,
	}, nil
}

// Available return the list migrations in the environment's directory
func (s Service) Available() ([]uint64, error) {
	fi, err := os.Stat(s.env.Directory)
	if err != nil {
		return nil, err
	}

	if !fi.Mode().IsDir() {
		return nil, fmt.Errorf("file %s isn't a directory", s.env.Directory)
	}

	raw, _ := filepath.Glob(filepath.Join(s.env.Directory, `*.sql`)) // The only possible error here is a pattern error

	var versions = make(map[uint64]struct{})
	for _, r := range raw {
		file := filepath.Base(r)

		chunks := strings.SplitN(file, `_`, 2)

		if len(chunks) != 2 {
			continue
		}

		version, err := strconv.ParseUint(chunks[0], 10, 64)
		if err != nil {
			continue
		}

		versions[version] = struct{}{}
	}

	var result []uint64
	for k, _ := range versions {
		result = append(result, k)
	}

	SortUint64s(result)

	return result, nil
}
