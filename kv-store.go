package main

type kvstore struct {
	mp map[string]string
}

// if value already exist, override it
func (kv *kvstore) SetValue(key, val string) {
	kv.mp[key] = val
}

func (kv *kvstore) GetValue(key string) string {
	val, ok := kv.mp[key]
	if !ok {
		val = "NULL"
	}
	return val
}

func (kv *kvstore) DeleteKey(key string) {
	delete(kv.mp, key)
}
