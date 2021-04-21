package cache

type Setter interface {
	Set(key string, value interface{}) error
}

type Getter interface {
	Get(key string) (interface{}, error)
}

type SetterGetter interface {
	Setter
	Getter
}
