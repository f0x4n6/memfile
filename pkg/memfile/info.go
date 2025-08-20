package memfile

type Info struct {
	data *Data
}

func (i *Info) Name() string {
	return i.data.Name()
}

func (i *Info) Size() int64 {
	i.data.RLock()
	defer i.data.RUnlock()
	return int64(len(i.data.buf))
}

func (i *Info) Mode() fs.FileMode {
	return 0 // regular
}

func (i *Info) ModTime() time.Time {
	i.data.RLock()
	defer i.data.RUnlock()
	return i.data.mod
}

func (i *Info) IsDir() bool {
	return false
}

func (i *Info) Sys() any {
	return nil
}
