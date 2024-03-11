package rewrite

import (
	"io"
	"os"
)

func File(name string, rewrite func(io.Reader, io.Writer) error) (err error) {
	w, err := os.OpenFile(name+"~", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := w.Close(); cerr != nil && err == nil {
			err = cerr
		}
		if err != nil {
			_ = os.Remove(w.Name())
		}
	}()
	r, err := os.Open(name)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := r.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	if err := rewrite(r, w); err != nil {
		return err
	}
	fs, err := r.Stat()
	if err != nil {
		return err
	}
	if err := w.Chmod(fs.Mode()); err != nil {
		return err
	}
	if err := w.Sync(); err != nil {
		return err
	}
	return os.Rename(w.Name(), name)
}
