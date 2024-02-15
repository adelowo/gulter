package gulter

import "mime/multipart"

type Option func(*Gulter)

func WithStorage(store Storage) Option {
	return func(gh *Gulter) {
		gh.storage = store
	}
}

func WithDestination(p string) Option {
	return func(gh *Gulter) {
		gh.destination = p
	}
}

func WithMaxFileSize(i int64) Option {
	return func(gh *Gulter) {
		gh.maxSize = i
	}
}

func WithFormFields(keys ...string) Option {
	return func(gh *Gulter) {
		gh.formKeys = append(gh.formKeys, keys...)
	}
}

func WithValidationFunc(validationFunc func(f multipart.File) error) Option {
	return func(g *Gulter) {
		g.validationFunc = validationFunc
	}
}

func WithNameFuncGenerator(nameFunc func(string) string) Option {
	return func(g *Gulter) {
		g.nameFuncGenerator = nameFunc
	}
}
