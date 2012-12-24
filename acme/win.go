// © 2012 Ethan Burns under the MIT license.

package acme

import (
	"code.google.com/p/goplan9/plan9/acme"
)

// Win is an acme window.
type Win struct {
	*acme.Win

	// Data implements the io.ReadWriter interface, reading and
	// writing the acme window's data file.
	Data winData
}

// New returns a new acme window.
func New(fmt string, args ...interface{}) (*Win, error) {
	w, err := acme.New()
	if err != nil {
		return nil, err
	}
	w.Name(fmt, args...)
	return &Win{
		Win:  w,
		Data: winData{w},
	}, nil
}

// winData implements io.ReadWriter using the acme window's data file.
type winData struct {
	*acme.Win
}

// Write implements the io.Writer interface, writing data in small
// chunks to the acme window's data file.
func (d winData) Write(data []byte) (int, error) {
	const maxWrite = 512
	var tot int
	for len(data) > 0 {
		sz := len(data)
		if sz > maxWrite {
			sz = maxWrite
		}
		n, err := d.Win.Write("data", data[:sz])
		tot += n
		if err != nil {
			return tot, err
		}
		data = data[n:]
	}
	return tot, nil
}

// Read implements the io.Reader interface, reading from the
// acme window's data file.
func (d winData) Read(data []byte) (int, error) {
	return d.Win.Read("data", data)
}
