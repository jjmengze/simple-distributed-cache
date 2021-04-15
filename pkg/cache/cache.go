package cache

type Setter interface {
	Set(key, value string) error
}

type Getter interface {
	Get(key string) (string, error)
}

type SetterGetter interface {
	Setter
	Getter
}
