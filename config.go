package gulter

type Option func(*Gulter)

func WithStorage(store Storage) Option {
	return func(gh *Gulter) {
		gh.storage = store
	}
}

// WithMaxFileSize allows you limit the size of file uploads to accept
func WithMaxFileSize(i int64) Option {
	return func(gh *Gulter) {
		gh.maxSize = i
	}
}

func WithValidationFunc(validationFunc ValidationFunc) Option {
	return func(g *Gulter) {
		g.validationFunc = validationFunc
	}
}

// WithNameFuncGenerator allows you configure how you'd like to rename your
// uploaded files
func WithNameFuncGenerator(nameFunc NameGeneratorFunc) Option {
	return func(g *Gulter) {
		g.nameFuncGenerator = nameFunc
	}
}

func WithIgnoreNonExistentKey(ignore bool) Option {
	return func(g *Gulter) {
		g.ignoreNonExistentKeys = ignore
	}
}
