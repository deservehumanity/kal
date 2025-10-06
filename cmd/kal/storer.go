package main

type Storer interface {
	Save([]Activity) error
	Load() ([]Activity, error)
}
