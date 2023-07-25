package api

import (
	"io/ioutil"
	"path/filepath"
)

func (r *Router) LoadRoutesFromFiles(files []string) error {
	for _, file := range files {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}

		if err := r.LoadRoutesFromJSON(data); err != nil {
			return err
		}
	}
	return nil
}

func (r *Router) LoadRoutesFromDir(dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		data, err := ioutil.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return err
		}

		if err := r.LoadRoutesFromJSON(data); err != nil {
			return err
		}
	}
	return nil
}
